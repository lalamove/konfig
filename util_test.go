package konfig

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUtils(t *testing.T) {
	var testCases = []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "MustGet",
			test: func(t *testing.T) {
				Set("test", true)
				require.Equal(t, true, MustGet("test"))
				require.Panics(t, func() {
					MustGet("foo")
				})
			},
		},
		{
			name: "IntSuccess",
			test: func(t *testing.T) {
				Set("foo", 1)
				var i = Int("foo")
				require.Equal(t, 1, i)
			},
		},
		{
			name: "MustIntSuccess",
			test: func(t *testing.T) {
				Set("foo", 1)
				var i int
				require.NotPanics(t, func() {
					i = MustInt("foo")
				})
				require.Equal(t, 1, i)
			},
		},
		{
			name: "MustIntPanic",
			test: func(t *testing.T) {
				require.Panics(t, func() {
					MustInt("foo")
				})
			},
		},
		{
			name: "StringSuccess",
			test: func(t *testing.T) {
				Set("foo", "bar")
				var s = String("foo")
				require.Equal(t, "bar", s)
			},
		},
		{
			name: "MustStringSuccess",
			test: func(t *testing.T) {
				Set("foo", "bar")
				var s string
				require.NotPanics(t, func() { s = MustString("foo") })
				require.Equal(t, "bar", s)
			},
		},
		{

			name: "MustStringPanics",
			test: func(t *testing.T) {
				require.Panics(t, func() { MustString("foo") })
			},
		},
		{

			name: "FloatSuccess",
			test: func(t *testing.T) {
				Set("foo", 2.1)
				var f = Float("foo")
				require.Equal(t, 2.1, f)
			},
		},
		{
			name: "MustFloatSuccess",
			test: func(t *testing.T) {
				Set("foo", 1.1)
				var f float64
				require.NotPanics(t, func() { f = MustFloat("foo") })
				require.Equal(t, 1.1, f)
			},
		},
		{
			name: "MustFloatPanics",
			test: func(t *testing.T) {
				require.Panics(t, func() { MustFloat("foo") })
			},
		},

		{
			name: "BoolSuccess",
			test: func(t *testing.T) {
				Set("foo", true)
				var b = Bool("foo")
				require.Equal(t, true, b)
			},
		},
		{
			name: "MustBoolSuccess",
			test: func(t *testing.T) {
				Set("foo", true)
				var b bool
				require.NotPanics(t, func() { b = MustBool("foo") })
				require.Equal(t, true, b)
			},
		},
		{
			name: "MustBoolPanics",
			test: func(t *testing.T) {
				require.Panics(t, func() { MustBool("foo") })
			},
		},

		{
			name: "DurationSuccess",
			test: func(t *testing.T) {
				Set("foo", "1m")
				var d = Duration("foo")
				require.Equal(t, 1*time.Minute, d)
			},
		},
		{
			name: "MustDurationSuccess",
			test: func(t *testing.T) {
				Set("foo", "1m")
				var d time.Duration
				require.NotPanics(t, func() { d = MustDuration("foo") })
				require.Equal(t, 1*time.Minute, d)
			},
		},
		{
			name: "MustDurationPanics",
			test: func(t *testing.T) {
				require.Panics(t, func() { MustDuration("foo") })
			},
		},
		{
			name: "TimeSuccess",
			test: func(t *testing.T) {
				Set("foo", "2019-01-02T15:04:05Z07:00")
				var d = Time("foo")

				var ti, _ = time.Parse(time.RFC3339, "2019-01-02T15:04:05Z07:00")

				require.Equal(t, ti, d)
			},
		},
		{
			name: "MustTimeSuccess",
			test: func(t *testing.T) {
				Set("foo", "2019-01-02T15:04:05Z07:00")
				var d time.Time

				var ti, _ = time.Parse(time.RFC3339, "2019-01-02T15:04:05Z07:00")

				require.NotPanics(t, func() { d = MustTime("foo") })

				require.Equal(t, ti, d)
			},
		},
		{
			name: "MustTimePanics",
			test: func(t *testing.T) {
				require.Panics(t, func() { MustTime("foo") })
			},
		},

		{
			name: "StringSliceSuccess",
			test: func(t *testing.T) {
				Set("foo", []string{"bar"})
				var b = StringSlice("foo")
				require.Equal(t, []string{"bar"}, b)
			},
		},
		{
			name: "MustStringSliceSuccess",
			test: func(t *testing.T) {
				Set("foo", []string{"bar"})
				var b []string
				require.NotPanics(t, func() { b = MustStringSlice("foo") })
				require.Equal(t, []string{"bar"}, b)
			},
		},
		{
			name: "MustStringSlicePanics",
			test: func(t *testing.T) {
				require.Panics(t, func() { MustStringSlice("foo") })
			},
		},
		{
			name: "IntSliceSuccess",
			test: func(t *testing.T) {
				Set("foo", []int{1})
				var b = IntSlice("foo")
				require.Equal(t, []int{1}, b)
			},
		},
		{
			name: "MustIntSliceSuccess",
			test: func(t *testing.T) {
				Set("foo", []int{1})
				var b []int
				require.NotPanics(t, func() { b = MustIntSlice("foo") })
				require.Equal(t, []int{1}, b)
			},
		},
		{
			name: "MustIntSlicePanics",
			test: func(t *testing.T) {
				require.Panics(t, func() { MustIntSlice("foo") })
			},
		},
		{
			name: "StringMapSuccess",
			test: func(t *testing.T) {
				Set("foo", map[string]interface{}{"foo": "bar"})
				var b = StringMap("foo")
				require.Equal(t, map[string]interface{}{"foo": "bar"}, b)
			},
		},
		{
			name: "MustStringMapSuccess",
			test: func(t *testing.T) {
				Set("foo", map[string]interface{}{"foo": "bar"})
				var b map[string]interface{}
				require.NotPanics(t, func() { b = MustStringMap("foo") })
				require.Equal(t, map[string]interface{}{"foo": "bar"}, b)
			},
		},
		{
			name: "MustStringMapPanics",
			test: func(t *testing.T) {
				require.Panics(t, func() { MustStringMap("foo") })
			},
		},
		{
			name: "StringMapStringSuccess",
			test: func(t *testing.T) {
				Set("foo", map[string]string{"foo": "bar"})
				var b = StringMapString("foo")
				require.Equal(t, map[string]string{"foo": "bar"}, b)
			},
		},
		{
			name: "MustStringMapStringSuccess",
			test: func(t *testing.T) {
				Set("foo", map[string]string{"foo": "bar"})
				var b map[string]string
				require.NotPanics(t, func() { b = MustStringMapString("foo") })
				require.Equal(t, map[string]string{"foo": "bar"}, b)
			},
		},
		{
			name: "MustStringMapStringPanics",
			test: func(t *testing.T) {
				require.Panics(t, func() { MustStringMapString("foo") })
			},
		},
		{
			name: "Exists",
			test: func(t *testing.T) {
				Set("test", true)
				require.True(t, Exists("test"))
				require.False(t, Exists("foo"))
			},
		},
	}
	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {
			reset()
			testCase.test(t)
		})
	}
}
