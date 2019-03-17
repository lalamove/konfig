package konfig

// Group gets a group of configs
func Group(groupName string) Store {
	return instance().Group(groupName)
}

// Group gets a group of configs
func (c *S) Group(groupName string) Store {
	return c.lazyGroup(groupName)
}

func (c *S) lazyGroup(groupName string) Store {
	c.mut.Lock()
	defer c.mut.Unlock()
	if v, ok := c.groups[groupName]; ok {
		return v
	}
	c.groups[groupName] = newStore(c.cfg)
	c.groups[groupName].name = groupName

	return c.groups[groupName]
}
