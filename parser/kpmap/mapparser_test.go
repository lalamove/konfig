package kpmap

import (
	"testing"

	"github.com/lalamove/konfig"
	"github.com/stretchr/testify/require"
)

func TestMapPopFlatten(t *testing.T) {
	var m = map[string]interface{}{
		"test": map[string]interface{}{
			"foo": "bar",
		},
		"testIface": map[interface{}]interface{}{
			1: "bar",
			"testIface": map[interface{}]interface{}{
				"foo": "bar",
			},
			"test": map[string]interface{}{
				"foo": "bar",
			},
		},
	}

	var v = konfig.Values{}
	PopFlatten(m, v)

	require.Equal(
		t,
		konfig.Values{
			"test.foo":                "bar",
			"testIface.1":             "bar",
			"testIface.testIface.foo": "bar",
			"testIface.test.foo":      "bar",
		},
		v,
	)
}
