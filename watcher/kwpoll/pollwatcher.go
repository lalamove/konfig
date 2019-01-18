package kwpoll

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-test/deep"
	"github.com/lalamove/konfig"
	"github.com/lalamove/nui/nlogger"
)

var (
	_ konfig.Watcher = (*PollWatcher)(nil)
	// ErrNoGetter is the error returned when no getter is set and Diff is set to true
	ErrNoGetter = errors.New("You must give a non nil getter to the poll diff watcher")
	// ErrAlreadyClosed is the error returned when trying to close an already closed PollDiffWatcher
	ErrAlreadyClosed = errors.New("PollDiffWatcher already closed")
)

// Rater is an interface that exposes a single
// Time method which returns the time until the next tick
type Rater interface {
	Time() time.Duration
}

// Getter is the interface to implement to fetch data to compare
type Getter interface {
	Get() (interface{}, error)
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
	Logger nlogger.Logger
	// Diff tells wether we should check for diffs
	// If diff is set, a Getter is required
	Diff bool
	// Getter is a getter to fetch data to check diff
	Getter Getter
	// InitValue is the initial value to compare with whe Diff is true
	InitValue interface{}
}

// PollWatcher is a konfig.Watcher that sends events every x time given in the konfig.
type PollWatcher struct {
	cfg       *Config
	err       error
	pv        interface{}
	watchChan chan struct{}
	done      chan struct{}
}

// New creates a new PollWatcher fromt the given config
func New(cfg *Config) *PollWatcher {
	if cfg.Diff && cfg.Getter == nil {
		panic(ErrNoGetter)
	}
	if cfg.Logger == nil {
		cfg.Logger = defaultLogger()
	}

	return &PollWatcher{
		cfg:       cfg,
		pv:        cfg.InitValue,
		done:      make(chan struct{}),
		watchChan: make(chan struct{}),
	}
}

// Done indicates wether the watcher is done or not
func (t *PollWatcher) Done() <-chan struct{} {
	return t.done
}

// Start starts the ticker watcher
func (t *PollWatcher) Start() error {
	if t.cfg.Debug {
		t.cfg.Logger.Debug(
			fmt.Sprintf(
				"Starting ticker watcher with rate: %d",
				t.cfg.Rater.Time()/time.Second,
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

	t.cfg.Logger.Debug(
		fmt.Sprintf(
			"Waiting rater duration: %v seconds",
			rate/time.Second,
		),
	)

	time.Sleep(rate)
	for {
		select {
		case <-t.done:
		default:
			if t.cfg.Debug {
				t.cfg.Logger.Debug("Tick")
			}
			if t.cfg.Diff {

				t.cfg.Logger.Debug(
					"Checking difference",
				)

				var r, err = t.cfg.Getter.Get()
				// We got error, we close
				if err != nil {
					t.cfg.Logger.Error(err.Error())
					t.err = err
					t.Close()
					return
				}

				if diff := deep.Equal(t.pv, r); diff != nil {
					if t.cfg.Debug {
						t.cfg.Logger.Debug(
							"Value is different: " + strings.Join(diff, "\n"),
						)
					}
					t.watchChan <- struct{}{}
					t.pv = r
				}

				t.cfg.Logger.Debug(
					"Value are the same, not updating",
				)
			} else {

				t.cfg.Logger.Debug(
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

func defaultLogger() nlogger.Logger {
	return nlogger.New(os.Stdout, "POLLWATCHER | ")
}
