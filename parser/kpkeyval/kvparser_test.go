package kpkeyval

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/lalamove/konfig"
	"github.com/stretchr/testify/require"
)

type ErrReader struct{}

func (e ErrReader) Read([]byte) (int, error) {
	return 0, errors.New("")
}

func TestKVParser(t *testing.T) {
	var testCases = []struct {
		name    string
		err     bool
		reader  io.Reader
		sep     string
		asserts func(t *testing.T, cfg konfig.Values)
	}{
		{
			name:   `no error, default separator`,
			err:    false,
			reader: strings.NewReader("BAR=FOO\nFOO=BAR"),
			asserts: func(t *testing.T, cfg konfig.Values) {
				require.Equal(t, "BAR", cfg["FOO"])
				require.Equal(t, "FOO", cfg["BAR"])
			},
		},
		{
			name:   `no error, custom separator`,
			err:    false,
			sep:    ":",
			reader: strings.NewReader("BAR:FOO\nFOO:BAR"),
			asserts: func(t *testing.T, cfg konfig.Values) {
				require.Equal(t, "BAR", cfg["FOO"])
				require.Equal(t, "FOO", cfg["BAR"])
			},
		},
		{
			name:   `err invalid format`,
			err:    true,
			sep:    ":",
			reader: strings.NewReader("BAR\nFOO"),
		},
		{
			name:   `err scanner`,
			err:    true,
			sep:    ":",
			reader: ErrReader{},
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				var v = konfig.Values{}
				var p = New(&Config{
					Sep: testCase.sep,
				})

				var err = p.Parse(testCase.reader, v)
				if testCase.err {
					require.NotNil(t, err, "err should not be nil")
					return
				}
				require.Nil(t, err, "err should be nil")

				testCase.asserts(t, v)
			},
		)
	}
}

func TestParserErr(t *testing.T) {
	var err = New(&Config{}).Parse(
		strings.NewReader(
			`invalid`,
		),
		konfig.Values{},
	)
	require.NotNil(t, err)
}
