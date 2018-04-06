package main

import (
	"bytes"
	"testing"
)

func benchmarkCrawl(baseUrl string, depth int, b *testing.B) {
	fetcher := SimpleFetcher{retries: 3, baseUrl: baseUrl}
	for i := 0; i < b.N; i++ {
		Crawl(baseUrl, depth, fetcher)
	}
}

func BenchmarkCrawlMonzo3(b *testing.B) { benchmarkCrawl("https://monzo.com", 3, b) }
func BenchmarkCrawlMonzo4(b *testing.B) { benchmarkCrawl("https://monzo.com", 6, b) }
func BenchmarkCrawlMonzo5(b *testing.B) { benchmarkCrawl("https://monzo.com", 10, b) }
func BenchmarkCrawlMonzo6(b *testing.B) { benchmarkCrawl("https://monzo.com", 6, b) }

/**
func BenchmarkCrawlGolang3(b *testing.B) { benchmarkCrawl("https://golang.org", 3, b) }
func BenchmarkCrawlGolang4(b *testing.B) { benchmarkCrawl("https://golang.org", 4, b) }
func BenchmarkCrawlGolang5(b *testing.B) { benchmarkCrawl("https://golang.org", 5, b) }
func BenchmarkCrawlGolang6(b *testing.B) { benchmarkCrawl("https://golang.org", 6, b) }
**/
func TestCrawl(t *testing.T) {
}

func TestStats(t *testing.T) {
	urlState := &uState{
		u: map[string]state{
			"u1": LOADED,
			"u2": LOADED,
			"u3": LOADED,
			"u4": LOADING,
			"u5": ERROR,
			"u6": LOADED,
			"u7": ERROR,
		},
	}
	suc, err := urlState.Stats()
	if suc != 4 || err != 3 {
		t.Error("Expected 4 fetches & 3 errors; got ",
			suc, " fetches & ", err, " errors")
	}
}

func TestPrettyPrintBuffer(t *testing.T) {
	collectedUrls := &urlCache{
		res: map[string][]string{
			"u0": {"u1", "u2", "u3"},
			"u1": {"u2", "u3"},
			"u2": {"u1", "u3"},
			"u3": {"u1", "u2"},
		},
	}
	var buffer bytes.Buffer

	s1 := collectedUrls.PrettyPrintBuffer("u0", 0, 1, buffer)
	expected1 := "u0\n"
	if s1.String() != expected1 {
		t.Error("\nExpected\n", expected1,
			"\nGot\n", s1.String())

	}
	s2 := collectedUrls.PrettyPrintBuffer("u0", 0, 2, buffer)
	expected2 :=
		`u0
└──u1
└──u2
└──u3`
	if s2.String() != expected2 {
		t.Error("\nExpected\n", expected2,
			"\nGot\n", s2.String())
	}
}
