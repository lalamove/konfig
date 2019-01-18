package konfig

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetter(t *testing.T) {
	t.Run(
		"test new getter",
		func(t *testing.T) {
			Init(DefaultConfig())

			var c = instance()
			c.Set("int", 1)

			var g = Getter("int")

			require.Equal(t, "1", g.String())
		},
	)
}
