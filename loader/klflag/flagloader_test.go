package klflag

import (
	"flag"
	"strings"
	"testing"
	"time"

	"github.com/lalamove/konfig"
	"github.com/stretchr/testify/require"
)

func TestFlagLoader(t *testing.T) {
	t.Run(
		"multiple flags",
		func(t *testing.T) {
			konfig.Init(konfig.DefaultConfig())

			var f = flag.NewFlagSet("foo", flag.ContinueOnError)
			f.Bool("foo", true, "")

			var loader = New(&Config{
				FlagSet: f,
			})

			var v = konfig.Values{}

			loader.Load(v)
			require.Equal(t, "true", v["foo"])
			require.Equal(t, defaultName, loader.Name())
		},
	)

	t.Run(
		"with replacer and prefix",
		func(t *testing.T) {
			konfig.Init(konfig.DefaultConfig())

			var fs = flag.NewFlagSet("foo", flag.ContinueOnError)
			fs.Bool("foo", true, "usage")

			var loader = New(&Config{
				Prefix:   "foo_",
				Replacer: strings.NewReplacer("foo", "bar"),
				FlagSet:  fs,
			})

			var v = konfig.Values{}

			loader.Load(v)
			require.Equal(t, "true", v["foo_bar"])
		},
	)

	t.Run(
		"default flag set",
		func(t *testing.T) {
			var loader = New(&Config{})
			require.True(t, loader.cfg.FlagSet == flag.CommandLine)
		},
	)

	t.Run(
		"max retry retry delay stop on failure",
		func(t *testing.T) {
			var loader = New(&Config{StopOnFailure: true, MaxRetry: 1, RetryDelay: 10 * time.Second})

			require.True(t, loader.StopOnFailure())
			require.Equal(t, 1, loader.MaxRetry())
			require.Equal(t, 10*time.Second, loader.RetryDelay())
		},
	)
}
