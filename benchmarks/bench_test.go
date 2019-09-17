package benchmarks

import (
	"testing"

	"github.com/lalamove/konfig"
	config "github.com/micro/go-micro/config"
	"github.com/micro/go-micro/config/source/memory"
	"github.com/spf13/viper"
)

func BenchmarkGetKonfig(b *testing.B) {
	var k = konfig.New(konfig.DefaultConfig())
	k.Set("foo", "bar")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k.Get("foo")
	}
}

func BenchmarkStringKonfig(b *testing.B) {
	var k = konfig.New(konfig.DefaultConfig())
	k.Set("foo", "bar")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k.String("foo")
	}
}

func BenchmarkGetViper(b *testing.B) {
	var v = viper.New()
	v.Set("foo", "bar")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Get("foo")
	}

}

func BenchmarkStringViper(b *testing.B) {
	var v = viper.New()
	v.Set("foo", "bar")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.GetString("foo")
	}
}

var data = []byte(`{
    "foo": "bar"
}`)

func newGoConfig() config.Config {
	memorySource := memory.NewSource(
		memory.WithJSON(data),
	)
	// Create new config
	conf := config.NewConfig()
	// Load file source
	conf.Load(memorySource)

	return conf
}

func BenchmarkGetGoConfig(b *testing.B) {
	conf := newGoConfig()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conf.Get("foo")
	}
}

func BenchmarkStringGoConfig(b *testing.B) {
	conf := newGoConfig()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conf.Get("foo").String("bar")
	}
}
