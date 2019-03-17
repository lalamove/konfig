package konfig

import (
	"fmt"
	"time"

	"github.com/spf13/cast"
)

type s map[string]interface{}

// exists returns a boolean indicating if the key exists in the map
func (m s) exists(k string) bool {
	_, ok := m[k]
	return ok
}

// checkStrictKeys checks if the given keys are present in the map. If a key is not
// found checkStrictKeys returns a non nil error.
func (m s) checkStrictKeys(keys []string) error {
	for _, k := range keys {
		if !m.exists(k) {
			return fmt.Errorf(ErrStrictKeyNotFoundMsg, k)
		}
	}
	return nil
}

// Exists checks if a config key k is set in the Store
func Exists(k string) bool {
	return instance().Exists(k)
}

// Exists checks if a config key k is set in the Store
func (c *S) Exists(k string) bool {
	var m = c.m.Load().(s)
	_, ok := m[k]
	return ok
}

// Get will return the value in config with given key k
// If not value is found, Get it returns nil
func Get(k string) interface{} {
	return instance().Get(k)
}

// MustGet returns the value in config with given key k
// If not found it panics
func MustGet(k string) interface{} {
	return instance().MustGet(k)
}

// Set will set the key value to the sync.Map
func Set(k string, v interface{}) {
	instance().Set(k, v)
}

// Set sets a value in config
func (c *S) Set(k string, v interface{}) {
	c.mut.Lock()
	defer c.mut.Unlock()

	var m = c.m.Load().(s)

	var nm = make(s)
	for kk, vv := range m {
		nm[kk] = vv
	}
	nm[k] = v

	// if there is a value bound we set it there also
	if c.v != nil {
		c.v.set(k, v)
	}

	c.m.Store(nm)
}

// Get gets a value from config
func (c *S) Get(k string) interface{} {
	var m = c.m.Load().(s)
	if v, ok := m[k]; ok {
		return v
	}
	return nil
}

// MustGet gets a value from config and panics if the value does not exist
func (c *S) MustGet(k string) interface{} {
	var m = c.m.Load().(s)
	if v, ok := m[k]; ok {
		return v
	}
	panic(fmt.Errorf(ErrConfigNotFoundMsg, k))
}

// MustInt gets the config k and tries to convert it to an int
// it panics if the config does not exist or it fails to convert it to an int.
func MustInt(k string) int {
	return instance().MustInt(k)
}

// MustInt gets the config k and tries to convert it to an int
// it panics if the config does not exist or it fails to convert it to an int.
func (c *S) MustInt(k string) int {
	return cast.ToInt(c.MustGet(k))
}

// Int gets the config k and tries to convert it to an int
// It returns the zero value if it doesn't find the config.
func Int(k string) int {
	return instance().Int(k)
}

// Int gets the config k and tries to convert it to an int
// It returns the zero value if it doesn't find the config.
func (c *S) Int(k string) int {
	return cast.ToInt(c.Get(k))
}

// MustFloat gets the config k and tries to convert it to a float64
// it panics if it fails.
func MustFloat(k string) float64 {
	return instance().MustFloat(k)
}

// MustFloat gets the config k and tries to convert it to a float64
// it panics if it fails.
func (c *S) MustFloat(k string) float64 {
	return cast.ToFloat64(c.MustGet(k))
}

// Float gets the config k and tries to convert it to float64
// It returns the zero value if it doesn't find the config.
func Float(k string) float64 {
	return instance().Float(k)
}

// Float gets the config k and tries to convert it to float64
// It returns the zero value if it doesn't find the config.
func (c *S) Float(k string) float64 {
	return cast.ToFloat64(c.Get(k))
}

// MustString gets the config k and tries to convert it to a string
// it panics if it fails.
func MustString(k string) string {
	return instance().MustString(k)
}

// MustString gets the config k and tries to convert it to a string
// it panics if it fails.
func (c *S) MustString(k string) string {
	return cast.ToString(c.MustGet(k))
}

// String gets the config k and tries to convert it to a string
// It returns the zero value if it doesn't find the config.
func String(k string) string {
	return instance().String(k)
}

// String gets the config k and tries to convert it to a string
// It returns the zero value if it doesn't find the config.
func (c *S) String(k string) string {
	return cast.ToString(c.Get(k))
}

// MustBool gets the config k and tries to convert it to a bool
// it panics if it fails.
func MustBool(k string) bool {
	return instance().MustBool(k)
}

// MustBool gets the config k and tries to convert it to a bool
// it panics if it fails.
func (c *S) MustBool(k string) bool {
	return cast.ToBool(c.MustGet(k))
}

// Bool gets the config k and converts it to a bool.
// It returns the zero value if it doesn't find the config.
func Bool(k string) bool {
	return instance().Bool(k)
}

// Bool gets the config k and converts it to a bool.
// It returns the zero value if it doesn't find the config.
func (c *S) Bool(k string) bool {
	return cast.ToBool(c.Get(k))
}

// MustDuration gets the config k and tries to convert it to a duration
// it panics if it fails.
func MustDuration(k string) time.Duration {
	return instance().MustDuration(k)
}

// MustDuration gets the config k and tries to convert it to a duration
// it panics if it fails.
func (c *S) MustDuration(k string) time.Duration {
	return cast.ToDuration(c.MustGet(k))
}

// Duration gets the config k and converts it to a duration.
// It returns the zero value if it doesn't find the config.
func Duration(k string) time.Duration {
	return instance().Duration(k)
}

// Duration gets the config k and converts it to a duration.
// It returns the zero value if it doesn't find the config.
func (c *S) Duration(k string) time.Duration {
	return cast.ToDuration(c.Get(k))
}

// MustTime gets the config k and tries to convert it to a time.Time
// it panics if it fails.
func MustTime(k string) time.Time {
	return instance().MustTime(k)
}

// MustTime gets the config k and tries to convert it to a time.Time
// it panics if it fails.
func (c *S) MustTime(k string) time.Time {
	return cast.ToTime(c.MustGet(k))
}

// Time gets the config k and converts it to a time.Time.
// It returns the zero value if it doesn't find the config.
func Time(k string) time.Time {
	return instance().Time(k)
}

// Time gets the config k and converts it to a time.Time.
// It returns the zero value if it doesn't find the config.
func (c *S) Time(k string) time.Time {
	return cast.ToTime(c.Get(k))
}

// MustStringSlice gets the config k and tries to convert it to a []string
// it panics if it fails.
func MustStringSlice(k string) []string {
	return instance().MustStringSlice(k)
}

// MustStringSlice gets the config k and tries to convert it to a []string
// it panics if it fails.
func (c *S) MustStringSlice(k string) []string {
	return cast.ToStringSlice(c.MustGet(k))
}

// StringSlice gets the config k and converts it to a []string.
// It returns the zero value if it doesn't find the config.
func StringSlice(k string) []string {
	return instance().StringSlice(k)
}

// StringSlice gets the config k and converts it to a []string.
// It returns the zero value if it doesn't find the config.
func (c *S) StringSlice(k string) []string {
	return cast.ToStringSlice(c.Get(k))
}

// MustIntSlice gets the config k and tries to convert it to a []int
// it panics if it fails.
func MustIntSlice(k string) []int {
	return instance().MustIntSlice(k)
}

// MustIntSlice gets the config k and tries to convert it to a []int
// it panics if it fails.
func (c *S) MustIntSlice(k string) []int {
	return cast.ToIntSlice(c.MustGet(k))
}

// IntSlice gets the config k and converts it to a []int.
// it returns the zero value if it doesn't find the config.
func IntSlice(k string) []int {
	return instance().IntSlice(k)
}

// IntSlice gets the config k and converts it to a []int.
// it returns the zero value if it doesn't find the config.
func (c *S) IntSlice(k string) []int {
	return cast.ToIntSlice(c.Get(k))
}

// MustStringMap gets the config k and tries to convert it to a map[string]interface{}
// it panics if it fails.
func MustStringMap(k string) map[string]interface{} {
	return instance().MustStringMap(k)
}

// MustStringMap gets the config k and tries to convert it to a map[string]interface{}
// it panics if it fails.
func (c *S) MustStringMap(k string) map[string]interface{} {
	return cast.ToStringMap(c.MustGet(k))
}

// StringMap gets the config k and converts it to a map[string]interface{}.
// it returns the zero value if it doesn't find the config.
func StringMap(k string) map[string]interface{} {
	return instance().StringMap(k)
}

// StringMap gets the config k and converts it to a map[string]interface{}.
// it returns the zero value if it doesn't find the config.
func (c *S) StringMap(k string) map[string]interface{} {
	return cast.ToStringMap(c.Get(k))
}

// MustStringMapString gets the config k and tries to convert it to a map[string]string
// it panics if it fails.
func MustStringMapString(k string) map[string]string {
	return instance().MustStringMapString(k)
}

// MustStringMapString gets the config k and tries to convert it to a map[string]string
// it panics if it fails.
func (c *S) MustStringMapString(k string) map[string]string {
	return cast.ToStringMapString(c.MustGet(k))
}

// StringMapString gets the config k and converts it to a map[string]string.
// it returns the zero value if it doesn't find the config.
func StringMapString(k string) map[string]string {
	return instance().StringMapString(k)
}

// StringMapString gets the config k and converts it to a map[string]string.
// it returns the zero value if it doesn't find the config.
func (c *S) StringMapString(k string) map[string]string {
	return cast.ToStringMapString(c.Get(k))
}
