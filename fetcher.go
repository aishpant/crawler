package main

/*
 * The simple fetcher here does not return a set of parsed links. Duplicates can be
 * present. It is assumed that caller will take care of them.
 * Usage:
 *	fetcher := SimpleFetcher{retries: 1, baseUrl: "https://example.com"}
 *	fetcher.Fetch("https://example.com")
 * TODO: log errors
 */

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Fetcher interface {
	// Fetch returns a slice of URLs found on that page
	Fetch(url string) (urls []string, err error)
}

type SimpleFetcher struct {
	retries int
	baseUrl string // needed for normalising relative paths & ignoring external paths
}

func (fetcher SimpleFetcher) Fetch(url string) ([]string, error) {
	if res, ok := fetcher.start(url, fetcher.retries); ok {
		return res, nil
	}
	return nil, fmt.Errorf("not found: %s", url)
}

// a global client for keep-alive
var client *http.Client

func (fetcher SimpleFetcher) start(url string, retries int) ([]string, bool) {

	var response *http.Response
	var err error
	t := &http.Transport{
		Dial: (&net.Dialer{
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 60 * time.Second,
	}
	client = &http.Client{
		Transport: t,
	}
	for retries >= 0 {
		response, err = client.Get(url)
		if err == nil {
			break
		}
		fmt.Println(err)
		retries--
	}
	defer response.Body.Close()
	if validUrls, ok := fetcher.parsePage(response.Body); ok {
		return validUrls, true
	}
	return nil, false
}

// assuming utf-8 encoded valid HTML
func (fetcher SimpleFetcher) parsePage(body io.Reader) ([]string, bool) {

	var validUrls []string

	t := html.NewTokenizer(body)
	for {
		tokenType := t.Next()
		switch tokenType {
		case html.ErrorToken:
			if t.Err() == io.EOF {
				return validUrls, true
			} else {
				fmt.Println(t.Err())
				return nil, false
			}
		case html.StartTagToken:
			token := t.Token()
			if token.DataAtom.String() == "a" {
				if url, ok := fetcher.getHrefAttr(token); ok {
					validUrls = append(validUrls, url)
				}
			}

		}
	}
	return nil, false
}

func (fetcher SimpleFetcher) getHrefAttr(token html.Token) (string, bool) {
	for _, attr := range token.Attr {
		if attr.Key == "href" {
			if url, ok := fetcher.cleanUpUrl(attr.Val); ok {
				return url, true
			}
		}
	}
	return "", false
}

func (fetcher SimpleFetcher) cleanUpUrl(link string) (string, bool) {
	// remove self-loops
	if link == "/" || strings.Contains(link, "#") {
		return "", false
	}
	u, err := url.Parse(link)
	if err != nil {
		//log.Fatal(err)
		return "", false
	}
	base, _ := url.Parse(fetcher.baseUrl)
	// normalise all urls, relative or absolute
	newLink := base.ResolveReference(u)
	// remove external links
	if base.Host != newLink.Host {
		return "", false
	}
	return newLink.String(), true
}
