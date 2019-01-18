package kpyaml

import (
	"io"

	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/parser"
	"github.com/lalamove/konfig/parser/kpmap"
	yaml "gopkg.in/yaml.v2"
)

// Parser is the YAML Parser it implements parser.Parser
var Parser = parser.Func(func(r io.Reader, s konfig.Values) error {
	var dec = yaml.NewDecoder(r)

	var d = make(map[string]interface{})
	var err = dec.Decode(&d)
	if err != nil {
		return err
	}

	kpmap.PopFlatten(d, s)

	return nil
})
