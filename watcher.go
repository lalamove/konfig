package konfig

// Watcher is the interface implementing a config watcher.
// Config watcher trigger loaders. A file watcher or a simple
// Timer can be valid watchers.
type Watcher interface {
	// Start starts the watcher, it must no be blocking.
	Start() error
	// Done indicate whether the watcher is done or not
	Done() <-chan struct{}
	// Watch should block until an event unlocks it
	Watch() <-chan struct{}
	// Close closes the watcher, it returns a non nil error if it is already closed
	// or something prevents it from closing properly.
	Close() error
	// Err returns the error attached to the watcher
	Err() error
}

// NopWatcher is a nil watcher
type NopWatcher struct{}

var _ Watcher = NopWatcher{}

// Done returns an already closed channel
func (NopWatcher) Done() <-chan struct{} {
	var c = make(chan struct{})
	close(c)
	return c
}

// Watch implements a basic watch that waits forever
func (NopWatcher) Watch() <-chan struct{} {
	var c = make(chan struct{})
	return c
}

// Close is a noop that returns nil
func (NopWatcher) Close() error {
	return nil
}

//Err returns nil, because nothing can go wrong when you do nothing
func (NopWatcher) Err() error {
	return nil
}

// Start implements watcher interface and always returns a nil error
func (NopWatcher) Start() error {
	return nil
}

// Watch starts the watchers on loaders
func Watch() error {
	return instance().Watch()
}

// Watch starts the watchers on loaders
func (c *S) Watch() error {

	// if metrics are enabled, we register them in prometheus
	if c.cfg.Metrics {
		if err := c.registerMetrics(); err != nil {
			return err
		}
	}

	for _, wl := range c.WatcherLoaders {
		if err := wl.Start(); err != nil {
			return err
		}
		go c.watchLoader(wl, 1)
	}
	return nil
}
