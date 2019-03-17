package konfig

// LoaderWatcher is an interface that implements both loader and watcher
type LoaderWatcher interface {
	Loader
	Watcher
}

type loaderWatcher struct {
	Loader
	Watcher
	values      Values
	name        string
	s           *S
	metrics     *loaderMetrics
	loaderHooks LoaderHooks
}

// NewLoaderWatcher creates a new LoaderWatcher from a Loader and a Watcher
func NewLoaderWatcher(l Loader, w Watcher) LoaderWatcher {
	return &loaderWatcher{
		Loader:  l,
		Watcher: w,
	}
}

func (c *S) newLoaderWatcher(l Loader, w Watcher, loaderHooks LoaderHooks) *loaderWatcher {
	var lw = &loaderWatcher{
		Loader:      l,
		Watcher:     w,
		s:           c,
		loaderHooks: loaderHooks,
	}

	if c.cfg.Metrics {
		lw.setMetrics()
	}

	return lw
}
