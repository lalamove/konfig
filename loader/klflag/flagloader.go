package klflag

import (
	"flag"
	"time"

	"github.com/lalamove/konfig"
	"github.com/lalamove/nui/nstrings"
)

var _ konfig.Loader = (*Loader)(nil)

const defaultName = "flag"

// Config is the config for the Flag Loader
type Config struct {
	// Name is the name of the loader
	Name string
	// StopOnFailure tells whether a failure to load configs should closed the config and all registered closers
	StopOnFailure bool
	// FlagSet is the flag set from which to load flags in config
	// default value is flag.CommandLine
	FlagSet *flag.FlagSet
	// Prefix is the prefix to append before each flag to be added in the konfig.Store
	Prefix string
	// Replacer is a replacer to apply on flags to be added in the konfig.Store
	Replacer nstrings.Replacer
	// MaxRetry is the maximum number of times to retry
	MaxRetry int
	// RetryDelay is the delay between each retry
	RetryDelay time.Duration
}

// Loader is a loader for command line flags
type Loader struct {
	cfg *Config
}

// New creates a new Loader with the given Config cfg
func New(cfg *Config) *Loader {
	if cfg.FlagSet == nil {
		cfg.FlagSet = flag.CommandLine
	}

	if cfg.Name == "" {
		cfg.Name = defaultName
	}

	return &Loader{
		cfg: cfg,
	}
}

// Name returns the name of the loader
func (l *Loader) Name() string { return l.cfg.Name }

// Load implements konfig.Loader interface, it loads flags from the FlagSet given in config
// into the konfig.Store
func (l *Loader) Load(s konfig.Values) error {
	l.cfg.FlagSet.VisitAll(func(f *flag.Flag) {
		var n = f.Name
		if l.cfg.Replacer != nil {
			n = l.cfg.Replacer.Replace(n)
		}
		s.Set(l.cfg.Prefix+n, f.Value.String())
	})
	return nil
}

// MaxRetry implements the konfig.Loader interface, it returns the max number of times a Load can be retried
// if it fails
func (l *Loader) MaxRetry() int {
	return l.cfg.MaxRetry
}

// RetryDelay implements the konfig.Loader interface, is the delay between each retry
func (l *Loader) RetryDelay() time.Duration {
	return l.cfg.RetryDelay
}

// StopOnFailure returns whether a load failure should stop the config and the registered closers
func (l *Loader) StopOnFailure() bool {
	return l.cfg.StopOnFailure
}
