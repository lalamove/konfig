package klfile

import (
	"errors"
	"os"
	"time"

	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/parser"
	"github.com/lalamove/konfig/watcher/kwfile"
	"github.com/lalamove/nui/nfs"
	"github.com/lalamove/nui/nlogger"
)

var (
	_ konfig.Loader = (*Loader)(nil)
	// ErrNoFiles is the error thrown when trying to create a file loader with no files in config
	ErrNoFiles = errors.New("no files provided")
	// ErrNoParser is the error thrown when trying to create a file loader with no parser
	ErrNoParser = errors.New("no parser provided")
	// DefaultRate is the default polling rate to check files
	DefaultRate = 10 * time.Second
)

const (
	defaultName = "file"
)

// File is a file to load from
type File struct {
	// Path is the path to the file
	Path string
	// Parser is the parser used to parse file and add it to the config store
	Parser parser.Parser
}

// Config is the config for the file loader
type Config struct {
	// Name is the name of the loader
	Name string
	// StopOnFailure tells whether a failure to load configs should closed the config and all registered closers
	StopOnFailure bool
	// Files is the path to the files to load
	Files []File
	// MaxRetry is the maximum number of times load can be retried in config
	MaxRetry int
	// RetryDelay is the delay between each retry
	RetryDelay time.Duration
	// Debug sets the debug mode on the file loader
	Debug bool
	// Logger is the logger used to print messages
	Logger nlogger.Provider
	// Watch sets whether the fileloader should also watch be a konfig.Watcher
	Watch bool
	// Rate is the kwfile polling rate
	// Default is 10 seconds
	Rate time.Duration
}

// Loader is the structure representring a file loader.
// A file loader loads data from a file and stores it in the konfig.Store.
type Loader struct {
	*kwfile.FileWatcher
	cfg *Config
	fs  nfs.FileSystem
}

// New creates a new Loader fromt the Config cfg.
func New(cfg *Config) *Loader {
	if cfg.Files == nil || len(cfg.Files) == 0 {
		panic(ErrNoFiles)
	}
	// make sure all files have a parser
	for _, f := range cfg.Files {
		if f.Parser == nil {
			panic(ErrNoParser)
		}
	}
	if cfg.Logger == nil {
		cfg.Logger = defaultLogger()
	}

	if cfg.Name == "" {
		cfg.Name = defaultName
	}

	// create the watcher
	var fw *kwfile.FileWatcher
	if cfg.Watch {
		var filePaths = make([]string, len(cfg.Files))
		for i, f := range cfg.Files {
			filePaths[i] = f.Path
		}
		fw = kwfile.New(
			&kwfile.Config{
				Files:  filePaths,
				Rate:   cfg.Rate,
				Debug:  cfg.Debug,
				Logger: cfg.Logger,
			},
		)
	}

	return &Loader{
		FileWatcher: fw,
		cfg:         cfg,
		fs:          nfs.OSFileSystem{},
	}
}

// NewFileLoader returns a new file loader with the given name n, the parser p and the file paths filePaths
func NewFileLoader(n string, p parser.Parser, filePaths ...string) *Loader {
	var files = make([]File, len(filePaths))
	for i, fp := range filePaths {
		files[i] = File{
			Path:   fp,
			Parser: p,
		}
	}

	return New(&Config{
		Name:  n,
		Files: files,
		Rate:  DefaultRate,
	})
}

// WithWatcher adds a watcher to the Loader
func (f *Loader) WithWatcher() *Loader {
	var filePaths = make([]string, len(f.cfg.Files))
	for i, fi := range f.cfg.Files {
		filePaths[i] = fi.Path
	}
	var fw = kwfile.New(
		&kwfile.Config{
			Files:  filePaths,
			Rate:   f.cfg.Rate,
			Debug:  f.cfg.Debug,
			Logger: f.cfg.Logger,
		},
	)
	f.FileWatcher = fw

	return f
}

// Name returns the name of the loader
func (f *Loader) Name() string { return f.cfg.Name }

// MaxRetry implements konfig.Loader interface and returns the maximum number
// of time Load method can be retried
func (f *Loader) MaxRetry() int {
	return f.cfg.MaxRetry
}

// RetryDelay implements konfig.Loader interface and returns the delay between each retry
func (f *Loader) RetryDelay() time.Duration {
	return f.cfg.RetryDelay
}

// Load implements the konfig.Loader interface. It reads from the file and adds the data to the konfig.Store.
func (f *Loader) Load(cfg konfig.Values) error {
	for _, file := range f.cfg.Files {
		var fd, err = f.fs.Open(file.Path)
		if err != nil {
			return err
		}

		// we parse the file
		if err := file.Parser.Parse(fd, cfg); err != nil {
			fd.Close()
			return err
		}
		fd.Close()
	}
	return nil
}

// StopOnFailure returns whether a load failure should stop the config and the registered closers
func (f *Loader) StopOnFailure() bool {
	return f.cfg.StopOnFailure
}

func defaultLogger() nlogger.Provider {
	return nlogger.NewProvider(nlogger.New(os.Stdout, "FILEWATCHER | "))
}
