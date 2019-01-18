package kpmap

import (
	"fmt"

	"github.com/lalamove/konfig"
)

func traverseMapIface(m map[interface{}]interface{}, s konfig.Values, p string) {
	for k, v := range m {
		var ks = fmt.Sprintf("%v", k)
		switch vt := v.(type) {
		case map[string]interface{}:
			traverseMap(vt, s, p+ks+konfig.KeySep)
		case map[interface{}]interface{}:
			traverseMapIface(vt, s, p+ks+konfig.KeySep)
		default:
			s.Set(p+ks, v)
		}
	}
}

func traverseMap(m map[string]interface{}, s konfig.Values, p string) {
	for k, v := range m {
		switch vt := v.(type) {
		case map[string]interface{}:
			traverseMap(vt, s, p+k+konfig.KeySep)
		case map[interface{}]interface{}:
			traverseMapIface(vt, s, p+k+konfig.KeySep)
		default:
			s.Set(p+k, v)
		}
	}
}

// PopFlatten populates a konfig.Store by flatteing a map[string]interface{}
func PopFlatten(m map[string]interface{}, s konfig.Values) {
	traverseMap(m, s, "")
}
