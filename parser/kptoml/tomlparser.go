package kptoml

import (
	"io"

	"github.com/BurntSushi/toml"
	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/parser"
	"github.com/lalamove/konfig/parser/kpmap"
)

// Parser parses the given json io.Reader and adds values in dot.path notation into the konfig.Store
var Parser = parser.Func(func(r io.Reader, s konfig.Values) error {
	// unmarshal the JSON into  map[string]interface{}
	var d = make(map[string]interface{})
	var _, err = toml.DecodeReader(r, &d)
	if err != nil {
		return err
	}

	kpmap.PopFlatten(d, s)

	return nil
})
