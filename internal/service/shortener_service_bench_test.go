package service

import "testing"

var benchURLs = []string{
	"https://practicum.yandex.ru/learn/go-advanced/courses/9b6f4ece-6c71-4b6f",
	"https://github.com/newmersedez/urlshort/internal/handler/handlers.go",
	"https://stackoverflow.com/questions/12345678/how-to-use-pprof-in-golang",
	"https://pkg.go.dev/net/http/pprof",
	"https://www.google.com/search?q=golang+memory+optimization",
}

func BenchmarkShorten(b *testing.B) {
	s := NewURLShortenerService()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Shorten(benchURLs[i%len(benchURLs)])
	}
}

func BenchmarkShortenAlloc(b *testing.B) {
	s := NewURLShortenerService()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Shorten(benchURLs[i%len(benchURLs)])
	}
}
