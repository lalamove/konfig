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
				Replacer: nstrings.ReplacerToLower,
			})

			var v = konfig.Values{}
			l.Load(v)

			require.Equal(t, "BAR", v["foo"])
			require.Equal(t, "FOO", v["bar"])
			require.Equal(t, defaultName, l.Name())
		},
	)

	t.Run(
		"load string slice values from defined env vars",
		func(t *testing.T) {
			os.Setenv("FOO", "BAR1,BAR2,BAR3")
			os.Setenv("BAR", "FOO") // we should get "string" value in store

			var l = New(&Config{
				Vars: []string{
					"FOO",
					"BAR",
				},
				SliceSeparator: ",",
			})

			var v = konfig.Values{}
			l.Load(v)

			require.Equal(t, []string{"BAR1", "BAR2", "BAR3"}, v["FOO"])
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
			require.False(t, ok)
		},
	)

	t.Run(
		"load string slice values from env vars regexp",
		func(t *testing.T) {
			konfig.Init(konfig.DefaultConfig())

			os.Setenv("FOO", "BAR1,BAR2,BAR3")
			os.Setenv("FAA", "VAL")
			os.Setenv("BAR", "FOO")

			var l = New(&Config{
				Regexp:         "^F.*",
				SliceSeparator: ",",
			})

			var v = konfig.Values{}
			l.Load(v)

			require.Equal(t, []string{"BAR1", "BAR2", "BAR3"}, v["FOO"])
			require.Equal(t, "VAL", v["FAA"])
			var _, ok = v["BAR"]
			require.False(t, ok)
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
			require.False(t, ok)
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
			require.False(t, ok)
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
