package klreader

import (
	"io"
	"time"

	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/parser"
)

const (
	defaultName = "reader"
)

// Config is the structure for the Loader's config
type Config struct {
	Name          string
	Reader        io.Reader
	Parser        parser.Parser
	MaxRetry      int
	RetryDelay    time.Duration
	StopOnFailure bool
}

// Loader is the Loader's structure
type Loader struct {
	cfg *Config
}

// New returns a new loader with the given config
func New(cfg *Config) *Loader {
	if cfg.Name == "" {
		cfg.Name = defaultName
	}
	return &Loader{
		cfg: cfg,
	}
}

// Name implements konfig.Loader. It returns the loader's name for metrics purpose.
func (l *Loader) Name() string { return l.cfg.Name }

// MaxRetry implements konfig.Loader interface and returns the maximum number
// of time Load method can be retried
func (l *Loader) MaxRetry() int {
	return l.cfg.MaxRetry
}

// RetryDelay implements konfig.Loader interface and returns the delay between each retry
func (l *Loader) RetryDelay() time.Duration {
	return l.cfg.RetryDelay
}

// Load implements the konfig.Loader interface. It reads from its io.Reader and adds the data to the konfig.Values
func (l *Loader) Load(cfg konfig.Values) error {
	return l.cfg.Parser.Parse(l.cfg.Reader, cfg)
}

// StopOnFailure returns whether a load failure should stop the config and the registered closers
func (l *Loader) StopOnFailure() bool {
	return l.cfg.StopOnFailure
}
