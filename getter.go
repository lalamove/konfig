package konfig

import "github.com/lalamove/nui/ngetter"

// Getter returns a mgetter.Getter for the key k
func Getter(k string) ngetter.GetterTyped {
	return instance().Getter(k)
}

// Getter returns a mgetter.Getter for the key k
func (c *S) Getter(k string) ngetter.GetterTyped {
	return ngetter.GetterTypedFunc(func() interface{} {
		return c.Get(k)
	})
}
