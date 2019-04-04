package parser

import (
	"io"

	"github.com/lalamove/konfig"
)

var _ Parser = (Func)(nil)

var _ Parser = (*NopParser)(nil)

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

// NopParser is a nil parser, useful for unit test
type NopParser struct {
	Err error
}

// Parse implements Parser interface
func (p NopParser) Parse(r io.Reader, s konfig.Values) error {
	return p.Err
}
