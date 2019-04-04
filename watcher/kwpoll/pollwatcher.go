package kwpoll

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/lalamove/konfig"
	"github.com/lalamove/nui/nlogger"
)

var (
	_ konfig.Watcher = (*PollWatcher)(nil)
	// ErrNoLoader is the error returned when no Loader is set and Diff is set to true
	ErrNoLoader = errors.New("You must give a non nil Loader to the poll diff watcher")
	// ErrAlreadyClosed is the error returned when trying to close an already closed PollDiffWatcher
	ErrAlreadyClosed = errors.New("PollDiffWatcher already closed")
	// ErrNoWatcherSupplied is the error returned when Watch in general config is false but a watcher is still being registered
	ErrNoWatcherSupplied = errors.New("watcher has to be supplied when registering a watcher")
	// defaultDuration is used in Rater if no Rater was supplied
	defaultDuration = time.Second * 5
)

// Rater is an interface that exposes a single
// Time method which returns the time until the next tick
type Rater interface {
	Time() time.Duration
}

// Time is a time.Duration which implements the Rater interface
type Time time.Duration

// Time returns the time.Duration
func (t Time) Time() time.Duration {
	return time.Duration(t)
}

// Config is the config of a PollWatcher
type Config struct {
	// Rater is the rater the PollWatcher calls to get the duration until the next tick
	Rater Rater
	// Debug sets the debug mode
	Debug bool
	// Logger is the logger used to log debug messages
	Logger nlogger.Provider
	// Diff tells whether we should check for diffs
	// If diff is set, a Getter is required
	Diff bool
	// Loader is a loader to fetch data to check diff
	Loader konfig.Loader
	// InitValue is the initial value to compare with whe Diff is true
	InitValue konfig.Values
}

// PollWatcher is a konfig.Watcher that sends events every x time given in the konfig.
type PollWatcher struct {
	cfg       *Config
	err       error
	pv        konfig.Values
	watchChan chan struct{}
	done      chan struct{}
}

// New creates a new PollWatcher fromt the given config
func New(cfg *Config) *PollWatcher {
	if cfg.Diff && cfg.Loader == nil {
		panic(ErrNoLoader)
	}
	if cfg.Logger == nil {
		cfg.Logger = defaultLogger()
	}
	if cfg.Rater == nil {
		cfg.Rater = defaultRater()
	}

	return &PollWatcher{
		cfg:       cfg,
		pv:        cfg.InitValue,
		done:      make(chan struct{}),
		watchChan: make(chan struct{}),
	}
}

// Done indicates whether the watcher is done or not
func (t *PollWatcher) Done() <-chan struct{} {
	return t.done
}

// Start starts the ticker watcher
func (t *PollWatcher) Start() error {
	if t == nil {
		panic(ErrNoWatcherSupplied)
	}

	if t.cfg.Debug {
		t.cfg.Logger.Get().Debug(
			fmt.Sprintf(
				"Starting ticker watcher with rate: %dms",
				t.cfg.Rater.Time()/time.Millisecond,
			),
		)
	}
	go t.watch()
	return nil
}

// Watch returns the channel to which events are written
func (t *PollWatcher) Watch() <-chan struct{} {
	return t.watchChan
}

// Err returns the poll watcher error
func (t *PollWatcher) Err() error {
	return t.err
}

func (t *PollWatcher) watch() {
	var rate = t.cfg.Rater.Time()

	t.cfg.Logger.Get().Debug(
		fmt.Sprintf(
			"Waiting rater duration: %dms",
			rate/time.Millisecond,
		),
	)

	time.Sleep(rate)
	for {
		select {
		case <-t.done:
		default:
			if t.cfg.Debug {
				t.cfg.Logger.Get().Debug("Tick")
			}
			if t.cfg.Diff {
				t.cfg.Logger.Get().Debug(
					"Checking difference",
				)

				var v = konfig.Values{}
				var err = t.cfg.Loader.Load(v)
				// We got error, we close
				if err != nil {
					t.cfg.Logger.Get().Error(err.Error())
					t.err = err
					t.Close()
					return
				}
				if !t.valuesEqual(v) {
					if t.cfg.Debug {
						t.cfg.Logger.Get().Debug(
							"Value is different: " + spew.Sdump(t.pv, v) + "\n",
						)
					}
					t.watchChan <- struct{}{}
					t.pv = v
				} else {
					t.cfg.Logger.Get().Debug(
						"Values are the same, not updating",
					)
				}
			} else {
				t.cfg.Logger.Get().Debug(
					"Sending watch event",
				)
				t.watchChan <- struct{}{}
			}
			time.Sleep(t.cfg.Rater.Time())
		}
	}
}

// Close closes the PollWatcher
func (t *PollWatcher) Close() error {
	select {
	case <-t.done:
		return ErrAlreadyClosed
	default:
		close(t.done)
	}
	return nil
}

func (t *PollWatcher) valuesEqual(v konfig.Values) bool {
	if len(v) != len(t.pv) {
		return false
	}

	for k, x := range v {
		if y, ok := t.pv[k]; ok {
			if y != x {
				return false
			}
			continue
		}
		return false
	}

	return true
}

func defaultLogger() nlogger.Provider {
	return nlogger.NewProvider(nlogger.New(os.Stdout, "POLLWATCHER | "))
}

func defaultRater() Rater {
	return Time(defaultDuration)
}
