package godi

import "testing"

func BenchmarkMakeSingleton(b *testing.B) {
	var di = New()
	di.BindSingleton("test", func(args ...interface{}) (interface{}, error) {
		return "test", nil
	})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		di.MustMake("test")
	}
}

func BenchmarkMake(b *testing.B) {
	var di = New()
	di.Bind("test", func(args ...interface{}) (interface{}, error) {
		return "test", nil
	})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		di.MustMake("test")
	}
}
