package main

import (
	"flag"
	"fmt"
	"strings"
	"sync"
)

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

	// crawl url
	foundUrls, err := fetcher.Fetch(url)

	urlState.Lock()
	if err != nil {
		// TODO: log this error
		urlState.setState(url, ERROR)
		urlState.Unlock()
		return
	}
	urlState.setState(url, LOADED)
	urlState.Unlock()

	collectedUrls.add(url, foundUrls)

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
	// Lock so only one goroutine at a time can access the map
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

var (
	baseUrl  string
	maxDepth int
)

func main() {

	flag.IntVar(&maxDepth, "depth", 3, "maximum depth to crawl upto")
	flag.StringVar(&baseUrl, "url", "https://golang.com/", "site to crawl")
	flag.Parse()

	// create a fetcher and pass it to the crawler
	fetcher := SimpleFetcher{retries: 3, baseUrl: baseUrl}
	Crawl(baseUrl, maxDepth, fetcher)

	// pretty print a tree like structure
	prettyPrint(baseUrl, 0)

}

var visited = make(map[string]bool)

func prettyPrint(baseUrl string, numTabs int) {
	fmt.Println(baseUrl)
	if v, ok := collectedUrls.res[baseUrl]; ok {
		nestedUrls := v
		if numTabs == maxDepth-1 {
			return
		}
		for _, url := range nestedUrls {
			if _, ok := visited[url]; url != baseUrl && !ok {
				fmt.Print(strings.Repeat("\t", numTabs))
				fmt.Print("└──")
				prettyPrint(url, numTabs+1)
				visited[baseUrl] = true
			}
		}
	}
	return
}
