package kpjson

import (
	"encoding/json"
	"io"

	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/parser"
	"github.com/lalamove/konfig/parser/kpmap"
)

// Parser parses the given json io.Reader and adds values in dot.path notation into the konfig.Store
var Parser = parser.Func(func(r io.Reader, s konfig.Values) error {
	// unmarshal the JSON into  map[string]interface{}
	var dec = json.NewDecoder(r)

	var d = make(map[string]interface{})
	var err = dec.Decode(&d)
	if err != nil {
		return err
	}

	kpmap.PopFlatten(d, s)

	return nil
})
