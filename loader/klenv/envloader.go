package klenv

import (
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/lalamove/konfig"
	"github.com/lalamove/nui/nstrings"
)

var (
	_ konfig.Loader = (*Loader)(nil)
)

const (
	sepEnvVar   = "="
	defaultName = "env"
)

// Config is the config a an EnvLoader
type Config struct {
	// Name is the name of the loader
	Name string
	// StopOnFailure tells whether a failure to load configs should closed the config and all registered closers
	StopOnFailure bool
	// Regexp will load the environment variable if it matches the given regexp
	Regexp string
	// Vars will load vars only present in the vars slice
	Vars []string
	// Prefix will add a prefix to the environment variables when adding them in the config store
	Prefix string
	// Replacer is used to replace chars in env vars keys
	Replacer nstrings.Replacer
	// MaxRetry is the maximum number of time the load method can be retried when it fails
	MaxRetry int
	// RetryDelay is the time betweel each retry
	RetryDelay time.Duration
	// SliceSeparator contains separator for values like `item1,item2,item3`.
	// Such values will be loaded as string slice if separator is not empty.
	SliceSeparator string
}

// Loader is the structure representing the environment loader
type Loader struct {
	cfg *Config
	r   *regexp.Regexp
}

// New return a new environment loader with the given config
func New(cfg *Config) *Loader {
	var r *regexp.Regexp
	if cfg.Regexp != "" {
		r = regexp.MustCompile(cfg.Regexp)
	}

	if cfg.Name == "" {
		cfg.Name = defaultName
	}

	return &Loader{
		cfg,
		r,
	}
}

// Name returns the name of the loader
func (l *Loader) Name() string { return l.cfg.Name }

func (l *Loader) convertValue(v string) interface{} {
	if l.cfg.SliceSeparator != "" {
		// do not load value as slice if it contains only one item
		// to avoid situation when all string values will be string slice values
		// binding mechanism correctly works when we try to bind one string to string slice
		if strings.Contains(v, l.cfg.SliceSeparator) {
			return strings.Split(v, l.cfg.SliceSeparator)
		}
	}

	return v
}

// Load implements konfig.Loader, it loads environment variables into the konfig.Store
// based on config passed to the loader
func (l *Loader) Load(s konfig.Values) error {
	if l.cfg.Vars != nil && len(l.cfg.Vars) > 0 {
		return l.loadVars(s)
	}
	for _, v := range os.Environ() {
		var spl = strings.SplitN(v, sepEnvVar, 2)
		// if has regex and key does not macth regexp we continue
		if l.r != nil && !l.r.MatchString(spl[0]) {
			continue
		}
		var k = spl[0]
		if l.cfg.Replacer != nil {
			k = l.cfg.Replacer.Replace(k)
		}
		k = l.cfg.Prefix + k
		s.Set(k, l.convertValue(spl[1]))
	}

	return nil
}

// MaxRetry returns the maximum number to retry a load when an error occurs
func (l *Loader) MaxRetry() int {
	return l.cfg.MaxRetry
}

// RetryDelay returns the delay between each load retry
func (l *Loader) RetryDelay() time.Duration {
	return l.cfg.RetryDelay
}

func (l *Loader) loadVars(s konfig.Values) error {
	for _, k := range l.cfg.Vars {
		var v = os.Getenv(k)
		if l.cfg.Replacer != nil {
			k = l.cfg.Replacer.Replace(k)
		}
		k = l.cfg.Prefix + k
		s.Set(k, l.convertValue(v))
	}
	return nil
}

// StopOnFailure returns whether a load failure should stop the config and the registered closers
func (l *Loader) StopOnFailure() bool {
	return l.cfg.StopOnFailure
}
