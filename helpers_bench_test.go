package skywalker

import "testing"

// BenchmarkRandomString measures the cost of generating a 32-character random
// string, a typical length for tokens and generated keys.
func BenchmarkRandomString(b *testing.B) {
	var s Skywalker

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.RandomString(32)
	}
}
