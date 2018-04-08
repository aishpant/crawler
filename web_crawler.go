package main

import (
	"flag"
	"fmt"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"strings"
	"sync"
	"time"
)

/*
 * This web crawler outputs a site map of crawled websites
 * Usage:
 *	./crawler -url https://golang.org/ -depth 4 -output out.txt
 * The sitemap is written to out.txt.
 *
 * This web crawler is inspired by the Web Crawler exercise from the gotour
 * https://tour.golang.org/concurrency/10
 */

func Crawl(url string, depth int, fetcher Fetcher) {

	if depth <= 0 {
		return
	}

	urlState.Lock()
	// return if url has already been crawled, or is being concurrently loaded
	if urlState.isLoaded(url) || urlState.isLoading(url) {
		urlState.Unlock()
		return
	}
	urlState.setState(url, LOADING)
	urlState.Unlock()

	foundUrls, err := fetcher.Fetch(url, HttpGetClient{})

	urlState.Lock()
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch url")
		urlState.setState(url, ERROR)
		urlState.Unlock()
		return
	}

	urlState.setState(url, LOADED)
	urlState.Unlock()

	collectedUrls.add(url, foundUrls)

	// make a channel per level
	ch := make(chan int)
	for _, u := range foundUrls {
		go func(u string) {
			Crawl(u, depth-1, fetcher)
			ch <- 1
		}(u)
	}

	// wait for concurrent tasks to finish
	for i := 0; i < len(foundUrls); i++ {
		<-ch
	}
	return
}

type urlCache struct {
	res map[string][]string
	sync.Mutex
}

func (cache *urlCache) add(url string, urls []string) {
	cache.Lock()
	cache.res[url] = urls
	cache.Unlock()
}

var collectedUrls = &urlCache{res: make(map[string][]string)}

type state int

const (
	LOADING state = iota
	LOADED
	ERROR
)

type uState struct {
	u map[string]state
	sync.Mutex
}

var urlState = &uState{u: make(map[string]state)}

func (urlState *uState) isLoaded(url string) bool {
	state, ok := urlState.u[url]
	if ok {
		return state == LOADED
	}
	return false
}

func (urlState *uState) isLoading(url string) bool {
	state, ok := urlState.u[url]
	if ok {
		return state == LOADING
	}
	return false
}

func (urlState *uState) setState(url string, state state) {
	urlState.u[url] = state
}

func main() {

	var (
		baseUrl    = flag.String("url", "https://golang.com/", "site to crawl")
		maxDepth   = flag.Int("depth", 3, "maximum depth to crawl upto")
		outputFile = flag.String("output", "out.txt", "file to write the sitemap to")
	)
	flag.Parse()

	// create a fetcher and pass it to the crawler
	fetcher := SimpleFetcher{retries: 3, baseUrl: *baseUrl}
	start := time.Now()
	Crawl(*baseUrl, *maxDepth, fetcher)
	elapsed := time.Since(start)
	log.Info().Msg(fmt.Sprintf("Crawl took %s\n", elapsed))

	start = time.Now()
	// pretty print a tree like structure
	s := collectedUrls.PrettyPrintBuffer(*baseUrl, 0, *maxDepth)
	err := ioutil.WriteFile(*outputFile, []byte(s), 0644)
	if err != nil {
		log.Fatal().Err(err).Msgf("could not write sitemap to file")
	}

	// print fetch stats
	success, errs := urlState.Stats()
	log.Info().Msgf("Found %v urls", success)
	log.Info().Msgf("%v Error(s) in fetch", errs)
	elapsed = time.Since(start)
	log.Info().Msgf("Pretty print took %s", elapsed)

}

func (collectedUrls *urlCache) PrettyPrintBuffer(baseUrl string, numTabs int, maxDepth int) string {

	if numTabs == maxDepth-1 {
		return baseUrl
	}

	var children string

	if nestedUrls, ok := collectedUrls.res[baseUrl]; ok {
		for _, url := range nestedUrls {
			nesting := strings.Repeat("\t", numTabs)
			join := "└──"
			child := "\n" + nesting + join + collectedUrls.PrettyPrintBuffer(url, numTabs+1, maxDepth)
			children += child
		}
	}
	return baseUrl + children
}

func (urlState *uState) Stats() (int, int) {
	success, err := 0, 0
	for _, state := range urlState.u {
		switch state {
		case LOADED:
			success++
		case ERROR, LOADING:
			err++
		}
	}
	return success, err
}
