package kpjson

import (
	"strings"
	"testing"

	"github.com/lalamove/konfig"
	"github.com/stretchr/testify/require"
)

func TestJSONParser(t *testing.T) {
	var testCases = []struct {
		name    string
		json    string
		asserts func(t *testing.T, s konfig.Values)
	}{
		{
			name: "simple 1 level json object",
			json: `{"foo":"bar","bar":"foo","int":1}`,
			asserts: func(t *testing.T, v konfig.Values) {
				require.Equal(
					t,
					"bar",
					v["foo"],
				)

				require.Equal(
					t,
					"foo",
					v["bar"],
				)

				require.Equal(
					t,
					float64(1),
					v["int"],
				)
			},
		},
		{
			name: "nested objects",
			json: `{
				"foo": "bar",
				"bar": {
					"foo": "hello world!",
					"bool": true,
					"nested": {
						"john": "doe"
					}
				}
			}`,
			asserts: func(t *testing.T, v konfig.Values) {
				require.Equal(
					t,
					"bar",
					v["foo"],
				)

				require.Equal(
					t,
					"hello world!",
					v["bar.foo"],
				)

				require.Equal(
					t,
					true,
					v["bar.bool"],
				)

				require.Equal(
					t,
					"doe",
					v["bar.nested.john"],
				)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				konfig.Init(konfig.DefaultConfig())

				var v = konfig.Values{}
				var err = Parser.Parse(
					strings.NewReader(
						testCase.json,
					),
					v,
				)

				require.Nil(t, err)
				testCase.asserts(t, v)
			},
		)
	}
}

func TestParserErr(t *testing.T) {
	var err = Parser.Parse(
		strings.NewReader(
			`invalid`,
		),
		konfig.Values{},
	)
	require.NotNil(t, err)
}
