package konfig

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jinzhu/copier"
	"github.com/spf13/cast"
)

const (
	// TagKey is the tag key to unmarshal config values to bound value
	TagKey = "konfig"
	// KeySep is the separator for config keys
	KeySep = "."
)

var (
	// ErrIncorrectValue is the error thrown when trying to bind an invalid type to a config store
	ErrIncorrectValue = errors.New("Bind takes a map[string]interface{} or a struct")
)

type value struct {
	s     *store
	v     *atomic.Value
	vt    reflect.Type
	mut   *sync.Mutex
	isMap bool
}

// Value returns the value bound to the root config store
func Value() interface{} {
	return instance().Value()
}

// Bind binds a value to the root config store
func Bind(v interface{}) {
	instance().Bind(v)
}

// Value returns the value bound to the config store
func (c *store) Value() interface{} {
	return c.v.v.Load()
}

// Bind binds a value (either a map[string]interface{} or a struct) to the config store. When config values are set on the config store, they are also set on the bound value.
func (c *store) Bind(v interface{}) {
	var t = reflect.TypeOf(v)
	var k = t.Kind()
	//  if it is neither a map nor a struct
	if k != reflect.Map && k != reflect.Struct {
		panic(ErrIncorrectValue)
	}
	// if it is a map check map[string]interface{}
	if k == reflect.Map &&
		(t.Key().Kind() != reflect.String || t.Elem().Kind() != reflect.Interface) {
		panic(ErrIncorrectValue)
	}

	var val = &value{
		s:     c,
		isMap: k == reflect.Map,
		mut:   &sync.Mutex{},
	}

	val.vt = t

	// create a new pointer to the given value and store it
	var atomicValue atomic.Value
	var n = reflect.Zero(val.vt)
	atomicValue.Store(n.Interface())

	val.v = &atomicValue

	c.v = val
}

func (val *value) set(k string, v interface{}) {
	val.mut.Lock()
	defer val.mut.Unlock()

	var configValue = val.v.Load()

	// if value is a map
	// store things in a map
	if val.isMap {
		var mapV = configValue.(map[string]interface{})
		var nMap = make(map[string]interface{})

		for kk, vv := range mapV {
			nMap[kk] = vv
		}

		nMap[k] = v

		val.v.Store(nMap)
		return
	}

	// make a copy
	var t = reflect.TypeOf(configValue)
	var nVal = reflect.New(t)

	copier.Copy(nVal.Interface(), configValue)

	val.setStruct(k, v, nVal.Interface())

	val.v.Store(nVal.Elem().Interface())
}

func (val *value) setValues(ox Values, x Values) {
	val.mut.Lock()
	defer val.mut.Unlock()

	var configValue = val.v.Load()

	// if value is a map
	// store things in a map
	if val.isMap {
		var mapV = configValue.(map[string]interface{})
		var nMap = make(map[string]interface{})

		for kk, vv := range mapV {
			nMap[kk] = vv
		}

		for kk, vv := range x {
			nMap[kk] = vv
		}

		val.v.Store(nMap)
		return
	}

	// make a copy
	var t = reflect.TypeOf(configValue)
	var nVal = reflect.New(t)

	copier.Copy(nVal.Interface(), configValue)

	// reset to zero value keys not present anymore
	for kk, vv := range ox {
		if _, ok := x[kk]; !ok {
			val.setStruct(
				kk,
				reflect.Zero(reflect.TypeOf(vv)).Interface(),
				nVal.Interface(),
			)
		}
	}

	for kk, vv := range x {
		val.setStruct(kk, vv, nVal.Interface())
	}
	val.v.Store(nVal.Elem().Interface())
}

func (val *value) setStruct(k string, v interface{}, targetValue interface{}) bool {
	// is a struct, find matching tag
	var valTypePtr = reflect.TypeOf(targetValue)
	var valType = valTypePtr.Elem()
	var valValuePtr = reflect.ValueOf(targetValue)
	var valValue = valValuePtr.Elem()
	var set bool

	for i := 0; i < valType.NumField(); i++ {
		var fieldType = valType.Field(i)
		var fieldName = fieldType.Name
		var tag = fieldType.Tag.Get(TagKey)

		// check tag, if it matches key
		// assign v to field
		if tag == k || strings.EqualFold(fieldName, k) {
			var field = valValue.FieldByName(fieldType.Name)
			if field.CanSet() {
				field.Set(reflect.ValueOf(castValue(field.Interface(), v)))
			}
			set = true
			continue

			// else if key has tag in prefix
		} else if strings.HasPrefix(k, tag+KeySep) ||
			strings.HasPrefix(strings.ToLower(k), strings.ToLower(fieldName)+KeySep) {

			var nK string

			if strings.HasPrefix(k, tag+KeySep) {
				nK = k[len(tag+KeySep):]
			} else {
				nK = k[len(fieldName+KeySep):]
			}

			switch fieldType.Type.Kind() {
			case reflect.Struct:
				var field = valValue.FieldByName(fieldType.Name)
				// if field can be set
				if field.CanSet() {
					var structType = field.Type()
					var nVal = reflect.New(structType)

					// we copy it
					copier.Copy(nVal.Interface(), field.Interface())

					// we set the field with the new struct
					if ok := val.setStruct(nK, v, nVal.Interface()); ok {
						field.Set(nVal.Elem())
						set = true
					}

					continue
				}
			case reflect.Ptr:
				if fieldType.Type.Elem().Kind() == reflect.Struct {
					var field = valValue.FieldByName(fieldType.Name)
					if field.CanSet() {
						var nVal = reflect.New(fieldType.Type.Elem())

						// if field is not nil
						// we copy it
						if !field.IsNil() {
							copier.Copy(nVal.Interface(), field.Interface())
						}

						if ok := val.setStruct(nK, v, nVal.Interface()); ok {
							field.Set(nVal)
							set = true
						}
						continue
					}
				}
			}
		}
	}

	if !set {
		val.s.cfg.Logger.Debug(
			fmt.Sprintf(
				"Config key %s not found in bound value",
				k,
			),
		)
	}

	return set
}

func castValue(f interface{}, v interface{}) interface{} {
	switch f.(type) {
	case string:
		return cast.ToString(v)
	case bool:
		return cast.ToBool(v)
	case int:
		return cast.ToInt(v)
	case int64:
		return cast.ToInt64(v)
	case int32:
		return cast.ToInt32(v)
	case float64:
		return cast.ToFloat64(v)
	case float32:
		return cast.ToFloat32(v)
	case uint64:
		return cast.ToUint64(v)
	case uint32:
		return cast.ToUint32(v)
	case uint8:
		return cast.ToUint8(v)
	case []string:
		return cast.ToStringSlice(v)
	case []int:
		return cast.ToIntSlice(v)
	case time.Time:
		return cast.ToTime(v)
	case time.Duration:
		return cast.ToDuration(v)
	case map[string]string:
		return cast.ToStringMapString(v)
	}
	return v
}
