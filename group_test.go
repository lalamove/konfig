package konfig

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGroup(t *testing.T) {
	t.Run(
		"test basic group",
		func(t *testing.T) {
			var c = newStore(DefaultConfig())
			var g = c.Group("test")

			c.Set("foo", "bar")
			require.Equal(t, nil, g.Get("foo"))

			g.Set("foo", "bar")
			var gg = c.Group("test")
			require.Equal(t, "bar", gg.Get("foo"))
		},
	)
}
