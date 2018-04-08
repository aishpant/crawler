package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Fetcher testing inspired from the Web Crawler exercise in the gotour
// https://tour.golang.org/concurrency/10

func BenchmarkCrawl(b *testing.B) {
	baseUrl := "https://golang.org"
	depth := 5
	fetcher := SimpleFetcher{retries: 3, baseUrl: baseUrl}
	for i := 0; i < b.N; i++ {
		Crawl(baseUrl, depth, fetcher)
	}
}

// create a fake fetcher
type fakeFetcher map[string]fakeResult
type fakeResult []string

var fetcher = fakeFetcher{
	"u0": fakeResult{"u1", "u2", "u3"},
	"u1": fakeResult{"u2", "u3"},
	"u2": fakeResult{"u1", "u3"},
	"u3": fakeResult{"u1", "u2"},
}

func (f fakeFetcher) Fetch(url string, client Client) ([]string, error) {
	if res, ok := f[url]; ok {
		return res, nil
	}
	return nil, fmt.Errorf("not found: %s", url)
}

// a fake list of fetched url cache
var fakeCollectedUrls = &urlCache{
	res: map[string][]string{
		"u0": {"u1", "u2", "u3"},
		"u1": {"u2", "u3"},
		"u2": {"u1", "u3"},
		"u3": {"u1", "u2"},
	},
}

func TestCrawl(t *testing.T) {

	// crawl from baseUrl u0 till depth 3, given a fake fetcher
	Crawl("u0", 3, fetcher)

	// comparing with a global value 'collectedUrls'; bad
	assert.Equal(t, fakeCollectedUrls, collectedUrls)
}

func TestPrettyPrintBuffer(t *testing.T) {

	s1 := fakeCollectedUrls.PrettyPrintBuffer("u0", 0, 1)
	expected1 := "u0"
	assert.Equal(t, expected1, s1)

	s2 := fakeCollectedUrls.PrettyPrintBuffer("u0", 0, 2)
	expected2 :=
		`u0
└──u1
└──u2
└──u3`
	assert.Equal(t, expected2, s2)

	s3 := fakeCollectedUrls.PrettyPrintBuffer("u0", 0, 3)
	expected3 :=
		`u0
└──u1
	└──u2
	└──u3
└──u2
	└──u1
	└──u3
└──u3
	└──u1
	└──u2`
	assert.Equal(t, expected3, s3)
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
