package konfig

import (
	"io"

	multierror "github.com/hashicorp/go-multierror"
)

// Closers is a multi closer
type Closers []io.Closer

// Close closes all closers in the multi closer and returns an error if an error was encountered.
// Error returned is multierror.Error. https://github.com/hashicorp/go-multierror
func (cs Closers) Close() error {
	var multiErr error
	for _, closer := range cs {
		if err := closer.Close(); err != nil {
			multierror.Append(multiErr, err)
		}
	}
	return multiErr
}
