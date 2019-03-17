package konfig

import (
	"reflect"

	"github.com/pkg/errors"
)

// Values is the values attached to a loader
type Values map[string]interface{}

// Set adds a key value to the Values
func (x Values) Set(k string, v interface{}) {
	x[k] = v
}

func (x Values) load(ox Values, c *S) ([]string, error) {
	c.mut.Lock()
	defer c.mut.Unlock()

	// load the previous key store
	var m = c.m.Load().(s)

	// we copy the previous store
	// but we omit what was on the previous values
	var updatedKeys = make([]string, 0, len(x))
	var nm = make(s)
	for kk, vv := range m {
		// if key is not in old loader values
		// add it to the new store
		// else if key is in new loader values but value is different or key is
		// in previous values but not in new ones
		// we add it to updatedKeys list
		if _, ok := ox[kk]; !ok {
			nm[kk] = vv
		} else if v, ok := x[kk]; (ok && !reflect.DeepEqual(v, vv)) || !ok { // value is different or not present anymore
			updatedKeys = append(updatedKeys, kk)
		}
	}
	// we add the new values
	for kk, vv := range x {
		nm[kk] = vv
		// if key is not present in old store, add it to updatedKeys as we are adding a new key in
		if _, ok := m[kk]; !ok {
			updatedKeys = append(updatedKeys, kk)
		}
	}

	// if we have strict keys setup on the store and we have already loaded configs
	// we check those keys now, if they are not present, we will return the error.
	if c.strictKeys != nil && c.loaded {
		if err := nm.checkStrictKeys(c.strictKeys); err != nil {
			err = errors.Wrap(err, "Error while checking strict keys")
			c.cfg.Logger.Get().Error(err.Error())
			return nil, err
		}
	}

	// if there is a value bound we set it there also
	if c.v != nil {
		c.v.setValues(nm)
	}

	// we didn't get any error, store the new config state
	c.m.Store(nm)

	return updatedKeys, nil
}
