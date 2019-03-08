package konfig

import "reflect"

// Values is the values attached to a loader
type Values map[string]interface{}

// Set adds a key value to the Values
func (x Values) Set(k string, v interface{}) {
	x[k] = v
}

func (x Values) load(ox Values, c *store) []string {
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

	// if there is a value bound we set it there also
	if c.v != nil {
		c.v.setValues(ox, x)
	}

	c.m.Store(nm)

	return updatedKeys
}
