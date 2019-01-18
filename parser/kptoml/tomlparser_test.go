package kptoml

import (
	"strings"
	"testing"

	"github.com/lalamove/konfig"
	"github.com/stretchr/testify/require"
)

func TestJSONParser(t *testing.T) {
	var testCases = []struct {
		name    string
		toml    string
		asserts func(t *testing.T, v konfig.Values)
	}{
		{
			name: "simple 1 level toml object",
			toml: `
foo = "bar"
bar = "foo"
int = 1
`,
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
					int64(1),
					v["int"],
				)
			},
		},
		{
			name: "nested objects",
			toml: `
foo = "bar"

[bar]   
foo = "hello world!"
bool = true

[bar.nested] 
john = "doe"
`,
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
						testCase.toml,
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
