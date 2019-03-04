package godi

import "testing"

func BenchmarkGet(b *testing.B) {
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
