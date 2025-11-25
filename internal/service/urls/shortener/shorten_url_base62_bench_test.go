package shortener

import (
	"math/big"
	"testing"
)

func BenchmarkShortenURLBase62(b *testing.B) {
	testURLs := []string{
		"https://example.com/very/long/url/path/that/needs/to/be/shortened",
		"https://www.google.com/search?q=benchmark+testing+in+go&oq=benchmark",
		"https://github.com/golang/go/wiki/Performance",
		"https://stackoverflow.com/questions/tagged/go",
		"https://golang.org/pkg/testing/#Benchmark",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shortenURLBase62(testURLs[i%len(testURLs)])
	}
}

func BenchmarkToBase62(b *testing.B) {
	// Создаем большое число для конвертации
	hash := [8]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	num := new(big.Int)
	num.SetBytes(hash[:])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		toBase62(num)
	}
}

func BenchmarkShortenURLBase62Parallel(b *testing.B) {
	testURLs := []string{
		"https://example.com/very/long/url/path/that/needs/to/be/shortened",
		"https://www.google.com/search?q=benchmark+testing+in+go&oq=benchmark",
		"https://github.com/golang/go/wiki/Performance",
		"https://stackoverflow.com/questions/tagged/go",
		"https://golang.org/pkg/testing/#Benchmark",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			shortenURLBase62(testURLs[i%len(testURLs)])
			i++
		}
	})
}
