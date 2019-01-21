package benchmarks

import (
	"testing"

	"github.com/lalamove/konfig"
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
