package klenv

import (
	"os"
	"testing"
	"time"

	"github.com/lalamove/konfig"
	"github.com/lalamove/nui/nstrings"
	"github.com/stretchr/testify/require"
)

func TestEnvLoader(t *testing.T) {
	t.Run(
		"load defined env vars",
		func(t *testing.T) {
			os.Setenv("FOO", "BAR")
			os.Setenv("BAR", "FOO")

			var l = New(&Config{
				Vars: []string{
					"FOO",
					"BAR",
				},
			})

			var v = konfig.Values{}
			l.Load(v)

			require.Equal(t, "BAR", v["FOO"])
			require.Equal(t, "FOO", v["BAR"])
		},
	)

	t.Run(
		"load env vars regexp",
		func(t *testing.T) {
			konfig.Init(konfig.DefaultConfig())

			os.Setenv("FOO", "BAR")
			os.Setenv("BAR", "FOO")

			var l = New(&Config{
				Regexp: "^F.*",
			})

			var v = konfig.Values{}
			l.Load(v)

			require.Equal(t, "BAR", v["FOO"])
			var _, ok = v["BAR"]
			require.Equal(t, false, ok)
		},
	)

	t.Run(
		"load env vars prefix regexp",
		func(t *testing.T) {
			konfig.Init(konfig.DefaultConfig())

			os.Setenv("FOO", "BAR")
			os.Setenv("BAR", "FOO")

			var l = New(&Config{
				Regexp: "^F.*",
				Prefix: "KONFIG_",
			})

			var v = konfig.Values{}
			l.Load(v)

			require.Equal(t, "BAR", v["KONFIG_FOO"])

			var _, ok = v["KONFIG_BAR"]
			require.Equal(t, false, ok)
		},
	)

	t.Run(
		"load env vars prefix regexp replacer",
		func(t *testing.T) {
			konfig.Init(konfig.DefaultConfig())

			os.Setenv("FOO", "BAR")
			os.Setenv("BAR", "FOO")

			var l = New(&Config{
				Regexp:   "^F.*",
				Prefix:   "KONFIG_",
				Replacer: nstrings.ReplacerToLower,
			})

			var v = konfig.Values{}
			l.Load(v)

			require.Equal(t, "BAR", v["KONFIG_foo"])

			var _, ok = v["KONFIG_bar"]
			require.Equal(t, false, ok)
		},
	)

	t.Run(
		"new loader invalid regexp",
		func(t *testing.T) {
			konfig.Init(konfig.DefaultConfig())

			os.Setenv("FOO", "BAR")
			os.Setenv("BAR", "FOO")

			require.Panics(t, func() {
				New(&Config{
					Regexp: "[",
				})
			})

		},
	)

	t.Run(
		"test max retry stop on failure",
		func(t *testing.T) {
			var l = New(&Config{
				MaxRetry:      1,
				RetryDelay:    1 * time.Second,
				StopOnFailure: true,
			})

			require.True(t, l.StopOnFailure())
			require.Equal(t, 1, l.MaxRetry())
			require.Equal(t, 1*time.Second, l.RetryDelay())
		},
	)
}
