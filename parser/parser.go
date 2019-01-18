package parser

import (
	"io"

	"github.com/lalamove/konfig"
)

var _ Parser = (Func)(nil)

// Parser is the interface to implement to parse a config file
type Parser interface {
	Parse(io.Reader, konfig.Values) error
}

// Func is a function implementing the Parser interface
type Func func(io.Reader, konfig.Values) error

// Parse implements Parser interface
func (f Func) Parse(r io.Reader, s konfig.Values) error {
	return f(r, s)
}
