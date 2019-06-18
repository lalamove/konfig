// Package kpkeyval provides a key value parser to parse an io.Reader's content
// of key/values with a configurable separator and add it into a konfig.Store.
package kpkeyval

import (
	"bufio"
	"errors"
	"io"
	"strings"

	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/parser"
)

// DefaultSep is the default key value separator
const DefaultSep = "="

// ErrInvalidConfigFileFormat is the error returned when a problem is encountered when parsing the
// config file
var (
	ErrInvalidConfigFileFormat = errors.New("Err invalid file format")
	// make sure Parser implements fileloader.Parser
	_ parser.Parser = (*Parser)(nil)
)

// Config is the configuration of the key value parser
type Config struct {
	// Sep is the separator between keys and values
	Sep string
}

// Parser implements fileloader.Parser
// It parses a file of key/values with a specific separator
// and stores in the konfig.Store
type Parser struct {
	cfg *Config
}

// New creates a new parser with the given config
func New(cfg *Config) *Parser {
	if cfg.Sep == "" {
		cfg.Sep = DefaultSep
	}
	return &Parser{
		cfg: cfg,
	}
}

// Parse implement the fileloader.Parser interface
func (k *Parser) Parse(r io.Reader, cfg konfig.Values) error {
	var scanner = bufio.NewScanner(r)
	for scanner.Scan() {
		var cfgKey = strings.Split(scanner.Text(), k.cfg.Sep)
		if len(cfgKey) < 2 {
			return ErrInvalidConfigFileFormat
		}
		cfg.Set(cfgKey[0], strings.Join(cfgKey[1:], k.cfg.Sep))
	}
	return scanner.Err()
}
