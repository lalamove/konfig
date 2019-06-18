package kwfile

import (
	"fmt"
	"os"
	"time"

	"github.com/lalamove/konfig"
	"github.com/lalamove/nui/nlogger"
	"github.com/radovskyb/watcher"
)

var _ konfig.Watcher = (*FileWatcher)(nil)
var defaultRate = 10 * time.Second

// Config is the config of a FileWatcher
type Config struct {
	// Files is the path to the files to watch
	Files []string
	// Rate is the rate at which the file is watched
	Rate time.Duration
	// Debug sets the debug mode on the filewatcher
	Debug bool
	// Logger is the logger used to print messages
	Logger nlogger.Provider
}

// FileWatcher watches over a file given in the config
type FileWatcher struct {
	cfg       *Config
	w         *watcher.Watcher
	err       error
	watchChan chan struct{}
}

// New creates a new FileWatcher from the given *Config cfg
func New(cfg *Config) *FileWatcher {
	if cfg.Logger == nil {
		cfg.Logger = defaultLogger()
	}
	if cfg.Rate == 0 {
		cfg.Rate = defaultRate
	}

	var w = watcher.New()

	for _, file := range cfg.Files {
		cfg.Logger.Get().Info("adding file to watch: " + file)
		if err := w.Add(file); err != nil {
			panic(err)
		}
	}

	return &FileWatcher{
		cfg:       cfg,
		w:         w,
		watchChan: make(chan struct{}),
	}
}

// Done indicates whether the filewatcher is done
func (fw *FileWatcher) Done() <-chan struct{} {
	return fw.w.Closed
}

// Start starts the file watcher
func (fw *FileWatcher) Start() error {
	go fw.watch()
	go func() error {
		if err := fw.w.Start(fw.cfg.Rate); err != nil {
			fw.cfg.Logger.Get().Error(err.Error())
			return err
		}
		return nil
	}()
	return nil
}

// Watch return the channel to which events are written
func (fw *FileWatcher) Watch() <-chan struct{} {
	return fw.watchChan
}

func (fw *FileWatcher) watch() {
	for {
		select {
		// we get an event, write to the struct chan
		// log if debug mode
		case e := <-fw.w.Event:
			if fw.cfg.Debug {
				fw.cfg.Logger.Get().Debug(fmt.Sprintf(
					"Event received %v",
					e,
				))
			}
			fw.watchChan <- struct{}{}
		case err := <-fw.w.Error:
			// log error
			fw.cfg.Logger.Get().Error(err.Error())
			fw.err = err
			fw.Close()
			return
		case <-fw.w.Closed:
			// watcher is closed, return
			return
		}
	}
}

// Close closes the FileWatcher
func (fw *FileWatcher) Close() error {
	fw.w.Close()
	return nil
}

// Err returns the file watcher error
func (fw *FileWatcher) Err() error {
	return fw.err
}

func defaultLogger() nlogger.Provider {
	return nlogger.NewProvider(nlogger.New(os.Stdout, "FILEWATCHER | "))
}
