package konfig

// Values is the values attached to a loader
type Values map[string]interface{}

// Set adds a key value to the Values
func (x Values) Set(k string, v interface{}) {
	x[k] = v
}

func (x Values) load(ox Values, c *store) {
	c.mut.Lock()
	defer c.mut.Unlock()

	var m = c.m.Load().(s)

	// we copy the previous store
	// but we omit what was on the previous values
	var nm = make(s)
	for kk, vv := range m {
		if _, ok := ox[kk]; !ok {
			nm[kk] = vv
		}
	}
	// we add the new values
	for kk, vv := range x {
		nm[kk] = vv
	}

	// if there is a value bound we set it there also
	if c.v != nil {
		c.v.setValues(ox, x)
	}

	c.m.Store(nm)
}
