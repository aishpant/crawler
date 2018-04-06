package main

import (
	"bytes"
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

	ch := make(chan bool)
	for _, u := range foundUrls {
		go func(u string) {
			Crawl(u, depth-1, fetcher)
			ch <- true
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

func main() {

	var (
		baseUrl  = flag.String("url", "https://golang.com/", "site to crawl")
		maxDepth = flag.Int("depth", 3, "maximum depth to crawl upto")
	)
	flag.Parse()

	// create a fetcher and pass it to the crawler
	fetcher := SimpleFetcher{retries: 3, baseUrl: *baseUrl}
	Crawl(*baseUrl, *maxDepth, fetcher)

	// pretty print a tree like structure
	var buffer bytes.Buffer
	s := collectedUrls.PrettyPrintBuffer(*baseUrl, 0, *maxDepth, buffer)
	fmt.Println(s.String())

	// print fetch stats
	success, err := urlState.Stats()
	fmt.Printf("Found %v urls\n", success)
	fmt.Printf("%v Error(s) in fetch", err)

}

func (collectedUrls *urlCache) PrettyPrintBuffer(baseUrl string, numTabs int, maxDepth int, buffer bytes.Buffer) []bytes.Buffer {
	fmt.Printf("key %s \n Buffer 1\n %s\n", baseUrl, buffer.String())
	buffer.WriteString(baseUrl)
	buffer.WriteString("\n")
	fmt.Println("Step 1\n", buffer.String())
	if numTabs == maxDepth-1 {
		fmt.Println("ret 1")
		return buffer
	}
	if nestedUrls, ok := collectedUrls.res[baseUrl]; ok {
		for _, url := range nestedUrls {
			//TODO: remove this check, move logic to fetcher
			fmt.Printf("url %s\n", url)
			if url != baseUrl {
				buffer.WriteString(strings.Repeat("\t", numTabs))
				buffer.WriteString("└──")
				fmt.Println("buffer arg")
				fmt.Println(buffer.String())
				buffer = collectedUrls.PrettyPrintBuffer(url, numTabs+1, maxDepth, buffer)
				fmt.Println("buffer updated")
				fmt.Println(buffer.String())

			}
		}
	}
	return buffer
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
