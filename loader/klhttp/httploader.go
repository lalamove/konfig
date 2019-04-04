package klhttp

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/parser"
	"github.com/lalamove/konfig/watcher/kwpoll"
)

var (
	defaultRate = 10 * time.Second
	// ErrNoSources is the error thrown when creating an Loader without sources
	ErrNoSources = errors.New("No sources provided")
)

const defaultName = "http"

// Client is the interface used to send the HTTP request.
// It is implemented by http.Client.
type Client interface {
	Do(*http.Request) (*http.Response, error)
}

// Source is an HTTP source and a Parser
type Source struct {
	URL    string
	Method string
	Body   io.Reader
	Parser parser.Parser
	// Prepare is a function to modify request before sending it
	Prepare func(*http.Request)
	// StatusCode is the status code expected from this source
	// If the status code of the response is different, an error is returned.
	// Default is 200.
	StatusCode int
}

// Config is the configuration of the Loader
type Config struct {
	// Name is the name of the loader
	Name string
	// StopOnFailure tells whether a failure to load configs should closed the config and all registered closers
	StopOnFailure bool
	// Sources is a list of remote sources
	Sources []Source
	// Client is the client used to fetch the file, default is http.DefaultClient
	Client Client
	// MaxRetry is the maximum number of retries when an error occurs
	MaxRetry int
	// RetryDelay is the delay between each retry
	RetryDelay time.Duration
	// Watch sets the whether changes should be watched
	Watch bool
	// Rater is the rater to pass to the poll write
	Rater kwpoll.Rater
	// Debug sets the debug mode
	Debug bool
}

// Loader loads a configuration remotely
type Loader struct {
	*kwpoll.PollWatcher
	cfg *Config
}

// New returns a new Loader with the given Config.
func New(cfg *Config) *Loader {
	if cfg.Client == nil {
		cfg.Client = http.DefaultClient
	}

	if cfg.Sources == nil || len(cfg.Sources) == 0 {
		panic(ErrNoSources)
	}

	if cfg.Name == "" {
		cfg.Name = defaultName
	}

	var l = &Loader{
		cfg: cfg,
	}

	for i, source := range cfg.Sources {
		if source.Method == "" {
			source.Method = http.MethodGet
		}
		cfg.Sources[i] = source
	}

	if cfg.Watch {
		var v = konfig.Values{}
		var err = l.Load(v)
		if err != nil {
			panic(err)
		}
		l.PollWatcher = kwpoll.New(&kwpoll.Config{
			Loader:    l,
			Rater:     cfg.Rater,
			InitValue: v,
			Diff:      true,
			Debug:     cfg.Debug,
		})
	}

	return l
}

// Name returns the name of the loader
func (r *Loader) Name() string { return r.cfg.Name }

// Load loads the config from sources and parses the response
func (r *Loader) Load(s konfig.Values) error {
	for _, source := range r.cfg.Sources {
		if b, err := source.Do(r.cfg.Client); err == nil {
			if err := source.Parser.Parse(b, s); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

// MaxRetry returns the MaxRetry config property, it implements the konfig.Loader interface
func (r *Loader) MaxRetry() int {
	return r.cfg.MaxRetry
}

// RetryDelay returns the RetryDelay config property, it implements the konfig.Loader interface
func (r *Loader) RetryDelay() time.Duration {
	return r.cfg.RetryDelay
}

// StopOnFailure returns whether a load failure should stop the config and the registered closers
func (r *Loader) StopOnFailure() bool {
	return r.cfg.StopOnFailure
}
