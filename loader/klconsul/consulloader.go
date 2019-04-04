package klconsul

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/parser"
	"github.com/lalamove/konfig/watcher/kwpoll"
	"github.com/lalamove/nui/nlogger"
	"github.com/lalamove/nui/nstrings"
)

var (
	defaultTimeout               = 5 * time.Second
	_              konfig.Loader = (*Loader)(nil)
)

const (
	defaultName = "consul"
)

// Key is an Consul Key to load
type Key struct {
	// Key is the consul key
	Key string
	// Parser is the parser for the key
	// If nil, the value is casted to a string before adding to the config.Store
	Parser parser.Parser
	// QueryOptions is the query options to pass when retrieving the key from consul
	QueryOptions *api.QueryOptions
}

// ConsulKV is an interface that consul client.KV implements. It is used to retrieve keys.
type ConsulKV interface {
	Get(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error)
}

// Config is the structure representing the config of a Loader
type Config struct {
	// Name is the name of the loader
	Name string
	// Client is the consul KV client
	Client *api.Client
	// StopOnFailure tells whether a load failure(after the retries) leads to closing the config and all registered closers
	StopOnFailure bool
	// Keys is the list of keys to fetch
	Keys []Key
	// Timeout is the timeout duration when fetching a key
	Timeout time.Duration
	// Prefix is a prefix to prepend keys when adding into the konfig.Store
	Prefix string
	// Replacer is a Replacer for the key before adding to the konfig.Store
	Replacer nstrings.Replacer
	// Watch tells if there should be a watcher with the loader
	Watch bool
	// Rater is the rater to pass to the poll watcher
	Rater kwpoll.Rater
	// MaxRetry is the maximum number of times we can retry to load if it fails
	MaxRetry int
	// RetryDelay is the time between each retry when a load fails
	RetryDelay time.Duration
	// Debug sets debug mode on the consulloader
	Debug bool
	// Logger is used across this package to produce logs
	Logger nlogger.Provider
	// StrictMode will raise error if key was not found
	// In false state, konfig will try to reload desired key(s)
	// up until they are not found
	StrictMode bool

	kvClient ConsulKV
}

// Loader is the structure of a loader
type Loader struct {
	*kwpoll.PollWatcher
	cfg *Config
}

// New returns a new loader with the given config
func New(cfg *Config) *Loader {
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}

	if cfg.Client == nil {
		panic(errors.New("no consul client was provided"))
	}

	if cfg.Logger == nil {
		cfg.Logger = defaultLogger()
	}

	if cfg.Name == "" {
		cfg.Name = defaultName
	}

	if cfg.kvClient == nil {
		cfg.kvClient = cfg.Client.KV()
	}

	var l = &Loader{
		cfg: cfg,
	}

	if cfg.Watch {
		var v = konfig.Values{}
		// we don't want to kill the process if there is an error
		// in the config
		if err := l.Load(v); err != nil {
			cfg.Logger.Get().Error(fmt.Sprintf("Can't read provided config: %v", err))
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
func (l *Loader) Name() string { return l.cfg.Name }

// Load implements konfig.Loader,
// it loads environment variables into the konfig.Store
// based on config passed to the loader
func (l *Loader) Load(s konfig.Values) error {
	for _, k := range l.cfg.Keys {
		kp, _, err := l.keyValue(k.Key)
		if err != nil {
			return err
		}
		if kp == nil && l.cfg.StrictMode {
			return fmt.Errorf("provided key \"%v\" was not found", k.Key)
		} else if kp == nil {
			l.cfg.Logger.Get().Warn(fmt.Sprintf("provided key \"%v\" was not found", k.Key))
			return nil
		}

		var configKey = l.cfg.Prefix + string(kp.Key)
		if l.cfg.Replacer != nil {
			configKey = l.cfg.Replacer.Replace(configKey)
		}

		// if the key has a parser, we parse the key value using the provided Parser
		// else we just convert the value to a string
		if k.Parser != nil {
			if err := k.Parser.Parse(bytes.NewReader(kp.Value), s); err != nil {
				return err
			}
		} else {
			s.Set(configKey, string(kp.Value))
		}
	}

	return nil
}

// MaxRetry is the maximum number of time to retry when a load fails
func (l *Loader) MaxRetry() int {
	return l.cfg.MaxRetry
}

// RetryDelay is the delay between each retry
func (l *Loader) RetryDelay() time.Duration {
	return l.cfg.RetryDelay
}

// StopOnFailure returns whether a load failure should stop the config and the registered closers
func (l *Loader) StopOnFailure() bool {
	return l.cfg.StopOnFailure
}

// keyValue is a quick helper to load KVPair from
// the consul server
func (l *Loader) keyValue(k string) (pair *api.KVPair, qm *api.QueryMeta, err error) {
	return l.cfg.kvClient.Get(k, nil)
}

func defaultLogger() nlogger.Provider {
	return nlogger.NewProvider(nlogger.New(os.Stdout, "CONSULLOADER | "))
}
