package main

import (
	"errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

/*
 * The simple fetcher returns a set of parsed links.
 * Usage:
 *	fetcher := SimpleFetcher{retries: 1, baseUrl: "https://example.com"}
 *	fetcher.Fetch("https://example.com", HttpGetClient{})
 */

type Client interface {
	Get(string) (*http.Response, error)
}

type HttpGetClient struct{}

func (h HttpGetClient) Get(url string) (*http.Response, error) {
	httpClient := &http.Client{}
	return httpClient.Get(url)
}

type Fetcher interface {
	// Fetch returns a slice of URLs found on that page
	Fetch(url string, client Client) (urls []string, err error)
}

type SimpleFetcher struct {
	retries int
	baseUrl string // needed for normalising relative paths & ignoring external paths
}

func (fetcher SimpleFetcher) Fetch(url string, client Client) ([]string, error) {

	var (
		response *http.Response
		err      error
	)

	retries := fetcher.retries
	for retries >= 0 {
		response, err = client.Get(url)
		if err == nil {
			break
		}
		retries--
	}

	if err != nil && retries < 0 {
		return nil, err
	}

	defer response.Body.Close()

	validUrls, err := fetcher.parsePage(url, response.Body)
	if err == nil {
		return validUrls, nil
	}

	return nil, err
}

// assuming utf-8 encoded valid HTML
func (fetcher SimpleFetcher) parsePage(parentUrl string, body io.Reader) ([]string, error) {

	var validUrls []string
	var urlSet = &stringSet{set: make(map[string]bool)}

	t := html.NewTokenizer(body)
	for {
		tokenType := t.Next()
		switch tokenType {
		case html.ErrorToken:
			if t.Err() == io.EOF {
				return validUrls, nil
			} else {
				log.Info().Msg(t.Err().Error())
			}
		case html.StartTagToken:
			token := t.Token()
			if url, ok := fetcher.GetAnchorHrefAttr(token); ok {
				if url, err := fetcher.CleanUpUrl(url, parentUrl); err == nil {
					if !urlSet.contains(url) {
						validUrls = append(validUrls, url)
						urlSet.add(url)
					}
				} else {
					log.Info().Msg(err.Error())
				}
			}
		}
	}
	return nil, errors.New("no urls found on page " + parentUrl)
}

func (fetcher SimpleFetcher) GetAnchorHrefAttr(token html.Token) (string, bool) {

	if token.DataAtom.String() == "a" {
		for _, attr := range token.Attr {
			if attr.Key == "href" {
				return attr.Val, true
			}
		}
	}
	return "", false
}

func (fetcher SimpleFetcher) CleanUpUrl(link string, parentLink string) (string, error) {

	if len(link) == 0 ||
		link == parentLink ||
		link == "/" ||
		strings.HasPrefix(link, "#") {
		return "", errors.New("invalid url type : " + link)
	}
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}
	base, _ := url.Parse(fetcher.baseUrl)
	// normalise all urls, relative or absolute
	newLink := base.ResolveReference(u)
	// remove external links
	if base.Host != newLink.Host {
		return "", errors.New("external urls are not crawled : " + link)
	}
	return strings.TrimSuffix(newLink.String(), "/"), nil
}

type stringSet struct {
	set map[string]bool
	sync.RWMutex
}

func (set *stringSet) add(value string) {
	set.Lock()
	set.set[value] = true
	set.Unlock()
}

func (set *stringSet) contains(value string) bool {
	set.RLock()
	defer set.RUnlock()
	_, found := set.set[value]
	return found
}
