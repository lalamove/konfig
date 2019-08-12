package konfig

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBind(t *testing.T) {
	t.Run(
		"panic invalid type",
		func(t *testing.T) {
			var s = Instance()
			require.Panics(t, func() { s.Bind(1) })
		},
	)

	t.Run(
		"valid type map",
		func(t *testing.T) {
			var s = Instance()
			var m = make(map[string]interface{})
			require.NotPanics(t, func() { s.Bind(m) })
		},
	)

	t.Run(
		"valid type struct",
		func(t *testing.T) {
			type testConfig struct {
				v string `konfig:"v"`
			}
			var s = Instance()
			var tc testConfig
			require.NotPanics(t, func() { s.Bind(tc) })
		},
	)

}

func TestBindStructStrict(t *testing.T) {
	t.Run(
		"panic type map",
		func(t *testing.T) {
			var s = Instance()
			var m = make(map[string]interface{})
			require.Panics(t, func() { s.BindStructStrict(m) })
		},
	)

	t.Run(
		"valid type struct",
		func(t *testing.T) {
			type testConfig struct {
				v string `konfig:"v"`
			}
			var s = Instance()
			var tc testConfig
			require.NotPanics(t, func() { s.BindStructStrict(tc) })
		},
	)

	t.Run(
		"should use konfig tag value",
		func(t *testing.T) {
			type testConfig struct {
				x string `konfig:"a"`
				y string `konfig:"b"`
				z string `konfig:"v"`
			}
			reset()
			Init(DefaultConfig())
			var tc testConfig
			BindStructStrict(tc)

			require.Equal(t, []string{"a", "b", "v"}, c.strictKeys)
		},
	)

	t.Run(
		"should use struct field name as keys",
		func(t *testing.T) {
			type testConfig struct {
				c string
				d string
				e string
			}
			reset()
			Init(DefaultConfig())
			var tc testConfig
			BindStructStrict(tc)

			require.Equal(t, []string{"c", "d", "e"}, c.strictKeys)
		},
	)

	t.Run(
		"should recursively add fields with struct type as keys",
		func(t *testing.T) {
			type testConfig struct {
				c string `konfig:"c"`
				d string
				e struct {
					a string `konfig:"b"`
					c string
				}
			}
			reset()
			Init(DefaultConfig())
			var tc testConfig
			BindStructStrict(tc)

			require.Equal(t, []string{"c", "d", "e.b", "e.c"}, c.strictKeys)
		},
	)

	t.Run(
		"should skip - field tag",
		func(t *testing.T) {
			type testConfig struct {
				a string `konfig:"a"`
				b string `konfig:"b"`
				v string `konfig:"-"`
			}
			reset()
			Init(DefaultConfig())
			var tc testConfig
			BindStructStrict(tc)

			require.Equal(t, []string{"a", "b"}, c.strictKeys)
		},
	)

}

func TestSetStruct(t *testing.T) {

	t.Run(
		"valid type struct",
		func(t *testing.T) {
			type TestConfigSub struct {
				VV  string  `konfig:"vv"`
				TT  int     `konfig:"tt"`
				B   bool    `konfig:"bool"`
				F   float64 `konfig:"float64"`
				U   uint64  `konfig:"uint64"`
				I64 int64   `konfig:"int64"`
			}
			type TestConfig struct {
				V          string        `konfig:"v"`
				T          TestConfigSub `konfig:"sub"`
				SubT       *TestConfigSub
				SubMapTPtr map[string]*TestConfigSub `konfig:"submaptptr"`
				SubMapT    map[string]TestConfigSub  `konfig:"submapt"`
			}

			var expectedConfig = TestConfig{
				V: "test",
				T: TestConfigSub{
					VV:  "test2",
					TT:  1,
					B:   true,
					F:   1.9,
					I64: 1,
				},
				SubT: &TestConfigSub{
					VV: "",
					TT: 2,
				},
				SubMapTPtr: map[string]*TestConfigSub{
					"foo": {
						VV: "woop",
						TT: 1,
					},
					"bar": {
						VV: "hello",
						TT: 1,
					},
				},
				SubMapT: map[string]TestConfigSub{
					"foo": {
						VV: "woop",
						TT: 1,
					},
					"bar": {
						VV: "hello",
						TT: 1,
					},
				},
			}

			Init(DefaultConfig())

			var tc TestConfig
			require.NotPanics(t, func() { Bind(tc) })

			// new values
			var v = Values{
				"v":                 "test",
				"sub.vv":            "test2",
				"sub.tt":            1,
				"subt.tt":           2,
				"sub.bool":          true,
				"sub.float64":       1.9,
				"sub.int64":         int64(1),
				"submapt.bar.vv":    "hello",
				"submapt.bar.tt":    1,
				"submapt.foo.vv":    "woop",
				"submapt.foo.tt":    1,
				"submaptptr.bar.vv": "hello",
				"submaptptr.bar.tt": 1,
				"submaptptr.foo.vv": "woop",
				"submaptptr.foo.tt": 1,
				"subt.notfound":     1,
			}

			v.load(Values{
				"v": "a",
			}, instance())

			var configValue = Value().(TestConfig)
			require.Equal(t, "test", configValue.V)
			require.Equal(t, "test2", configValue.T.VV)
			require.Equal(t, 1, configValue.T.TT)
			require.Equal(t, 2, configValue.SubT.TT)
			require.Equal(t, true, configValue.T.B)
			require.Equal(t, 1.9, configValue.T.F)
			require.Equal(t, int64(1), configValue.T.I64)
			require.Equal(
				t,
				&TestConfigSub{
					VV: "woop",
					TT: 1,
				},
				configValue.SubMapTPtr["foo"],
				"SubMapT['foo'] should be equal",
			)
			require.Equal(
				t,
				&TestConfigSub{
					VV: "hello",
					TT: 1,
				},
				configValue.SubMapTPtr["bar"],
				"SubMapT['bar'] should be equal",
			)
			require.Equal(
				t,
				TestConfigSub{
					VV: "woop",
					TT: 1,
				},
				configValue.SubMapT["foo"],
				"SubMapT['foo'] should be equal",
			)
			require.Equal(
				t,
				TestConfigSub{
					VV: "hello",
					TT: 1,
				},
				configValue.SubMapT["bar"],
				"SubMapT['bar'] should be equal",
			)

			require.Equal(t, expectedConfig, Value())

			var vv = Values{
				"v":      "test",
				"sub.vv": "test2",
			}

			vv.load(v, instance())

			configValue = Value().(TestConfig)
			require.Equal(t, "test", configValue.V)
			require.Equal(t, "test2", configValue.T.VV)
			require.Equal(t, 0, configValue.T.TT)
			require.Nil(t, configValue.SubT)
			require.Nil(t, configValue.SubMapT)
			require.Nil(t, configValue.SubMapTPtr)
		},
	)

	t.Run(
		"valid type struct",
		func(t *testing.T) {
			type TestConfigSub struct {
				VV string `konfig:"vv"`
				TT int    `konfig:"tt"`
			}
			type TestConfig struct {
				V    string        `konfig:"v"`
				T    TestConfigSub `konfig:"sub"`
				SubT *TestConfigSub
			}

			var expectedConfig = TestConfig{
				V: "test",
				T: TestConfigSub{
					VV: "bar",
					TT: 1,
				},
				SubT: &TestConfigSub{
					VV: "",
					TT: 2,
				},
			}

			Init(DefaultConfig())

			var tc TestConfig
			require.NotPanics(t, func() { Bind(tc) })

			Set("v", "test")
			Set("sub.vv", "bar")
			Set("sub.tt", 1)
			Set("subt.tt", 2)

			var configValue = Value().(TestConfig)
			require.Equal(t, "test", configValue.V)
			require.Equal(t, "bar", configValue.T.VV)
			require.Equal(t, 1, configValue.T.TT)
			require.Equal(t, 2, configValue.SubT.TT)

			require.Equal(t, expectedConfig, Value())

		},
	)

	t.Run(
		"valid type map",
		func(t *testing.T) {
			Init(DefaultConfig())

			var tc = make(map[string]interface{})
			require.NotPanics(t, func() { Bind(tc) })

			Set("v", "test")
			Set("sub.vv", "bar")
			Set("sub.tt", 1)
			Set("subt.tt", 2)

			var configValue = Value().(map[string]interface{})
			require.Equal(t, "test", configValue["v"])
			require.Equal(t, "bar", configValue["sub.vv"])
			require.Equal(t, 1, configValue["sub.tt"])
			require.Equal(t, 2, configValue["subt.tt"])
		},
	)

	t.Run(
		"valid type map",
		func(t *testing.T) {
			Init(DefaultConfig())

			var tc = make(map[string]interface{})
			require.NotPanics(t, func() { Bind(tc) })

			var v = Values{
				"v":       "test",
				"sub.vv":  "test2",
				"sub.tt":  1,
				"subt.tt": 2,
			}

			v.load(Values{}, instance())

			var configValue = Value().(map[string]interface{})
			require.Equal(t, "test", configValue["v"])
			require.Equal(t, "test2", configValue["sub.vv"])
			require.Equal(t, 1, configValue["sub.tt"])
			require.Equal(t, 2, configValue["subt.tt"])
		},
	)
}

func TestCastValue(t *testing.T) {
	var testCases = []struct {
		x         interface{}
		y         interface{}
		expectedV interface{}
	}{
		{
			x:         "",
			y:         1,
			expectedV: "1",
		},
		{
			x:         true,
			y:         "true",
			expectedV: true,
		},
		{
			x:         int32(1),
			y:         "1",
			expectedV: int32(1),
		},
		{
			x:         float32(1),
			y:         "1",
			expectedV: float32(1),
		},
		{
			x:         uint64(1),
			y:         "1",
			expectedV: uint64(1),
		},
		{
			x:         uint32(1),
			y:         "1",
			expectedV: uint32(1),
		},
		{
			x:         uint8(1),
			y:         "1",
			expectedV: uint8(1),
		},
		{
			x:         []string{},
			y:         []interface{}{1},
			expectedV: []string{"1"},
		},
		{
			x:         []int{},
			y:         []interface{}{"1"},
			expectedV: []int{1},
		},
		{
			x:         time.Time{},
			y:         "2015-01-01",
			expectedV: time.Date(2015, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			x:         time.Duration(1 * time.Millisecond),
			y:         "1ms",
			expectedV: time.Duration(1 * time.Millisecond),
		},
		{
			x:         map[string]string{},
			y:         map[interface{}]interface{}{"foo": "bar"},
			expectedV: map[string]string{"foo": "bar"},
		},
	}

	for _, testCase := range testCases {
		t.Run(
			fmt.Sprintf("%T", testCase.x),
			func(t *testing.T) {
				var v = castValue(testCase.x, testCase.y)
				require.Equal(t, testCase.expectedV, v)
			},
		)
	}
}
