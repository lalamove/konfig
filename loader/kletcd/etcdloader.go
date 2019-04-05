package kletcd

import (
	"bytes"
	"context"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/parser"
	"github.com/lalamove/konfig/watcher/kwpoll"
	"github.com/lalamove/nui/ncontext"
	"github.com/lalamove/nui/nstrings"
	"go.etcd.io/etcd/clientv3"
)

var (
	defaultTimeout               = 5 * time.Second
	_              konfig.Loader = (*Loader)(nil)
)

const (
	defaultName = "etcd"
)

// Key is an Etcd Key to load
type Key struct {
	// Key is the etcd key
	Key string
	// Parser is the parser for the key
	// If nil, the value is casted to a string before adding to the config.Store
	Parser parser.Parser
}

// Config is the structure representing the config of a Loader
type Config struct {
	// Name is the name of the loader
	Name string
	// StopOnFailure tells whether a failure to load configs should closed the config and all registered closers
	StopOnFailure bool
	// Client is the etcd client
	Client *clientv3.Client
	// Keys is the list of keys to fetch
	Keys []Key
	// Timeout is the timeout duration when fetching a key
	Timeout time.Duration
	// Prefix is a prefix to prepend keys when adding into the konfig.Store
	Prefix string
	// Replacer is a Replacer for the key before adding to the konfig.Store
	Replacer nstrings.Replacer
	// Watch tells whether there should be a watcher with the loader
	Watch bool
	// Rater is the rater to pass to the poll watcher
	Rater kwpoll.Rater
	// MaxRetry is the maximum number of times we can retry to load if it fails
	MaxRetry int
	// RetryDelay is the time between each retry when a load fails
	RetryDelay time.Duration
	// Debug sets debug mode on the etcdloader
	Debug bool
	// Contexter provides a context, default value is contexter wrapping context package. It is used mostly for testing.
	Contexter ncontext.Contexter

	kvClient clientv3.KV
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

	if cfg.Contexter == nil {
		cfg.Contexter = ncontext.DefaultContexter
	}

	if cfg.Name == "" {
		cfg.Name = defaultName
	}

	if cfg.kvClient == nil {
		cfg.kvClient = cfg.Client.KV
	}

	var l = &Loader{
		cfg: cfg,
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
			Debug:     cfg.Debug,
			Diff:      true,
		})
	}

	return l
}

// Name returns the name of the loader
func (l *Loader) Name() string { return l.cfg.Name }

// Load loads the values from the keys defined by the config in the konfig.Store
func (l *Loader) Load(s konfig.Values) error {
	for _, k := range l.cfg.Keys {

		values, err := l.keyValue(k.Key)
		if err != nil {
			return err
		}

		for _, v := range values {
			var configKey = l.cfg.Prefix + string(v.Key)
			if l.cfg.Replacer != nil {
				configKey = l.cfg.Replacer.Replace(configKey)
			}

			// if the key has a parser, we parse the key value using the provided Parser
			// else we just convert the value to a string
			if k.Parser != nil {
				if err := k.Parser.Parse(bytes.NewReader(v.Value), s); err != nil {
					return err
				}
			} else {
				s.Set(configKey, string(v.Value))
			}
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

func (l *Loader) keyValue(k string) ([]*mvccpb.KeyValue, error) {
	var ctx, cancel = l.cfg.Contexter.WithTimeout(
		context.Background(),
		l.cfg.Timeout,
	)
	defer cancel()

	values, err := l.cfg.kvClient.Get(ctx, k)
	if err != nil {
		return nil, err
	}

	return values.Kvs, nil
}

// StopOnFailure returns whether a load failure should stop the config and the registered closers
func (l *Loader) StopOnFailure() bool {
	return l.cfg.StopOnFailure
}
