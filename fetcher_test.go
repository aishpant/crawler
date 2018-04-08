package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"testing"
)

var mockSimpleFetcher = SimpleFetcher{retries: 1, baseUrl: "https://example.com"}

// we assume the standard http client works well and just test the fetcher
// properties - uniqueness of links and url clean-ups
type fakeHttpClient struct{}

func (f fakeHttpClient) Get(url string) (*http.Response, error) {

	// a fake http get response
	htmlBody := `<!DOCTYPE html>
		<head>
		  <title>Blog • Aishwarya Pant</title>
		</head>

		<body class='section type-blog'>
		<div class='site'>
		    <a class='screen-reader' href='#main'>Skip to Content</a>
		    <header id='header' class='header-container'>
		      <div class='header site-header'>
			<nav id='main-menu' class='main-menu-container' aria-label='Main Menu'>
		  <ul class='main-menu'>
		  <li>
		      <a href='/'>Home</a>
		    </li>
		  <li>
		      <a class='current' aria-current='page' href='/blog/'>blog</a>
		    </li>
		  <li>
		      <a href='/about/'>About</a>
		    </li>
			<li>
		      <a href='/blog'>About</a>
		    </li>
			<li>
		      <a href='/start-outreachy'>You should apply for Outreachy!</a>
		    </li>
		  </ul>
		</nav>
		    </header>
		/body>
		</html>`

	// NopCloser implments Reader and Closer, we can send out faked response
	// here
	resp := &http.Response{
		Body: ioutil.NopCloser(bytes.NewBufferString(htmlBody)),
	}

	return resp, nil
}

func TestFetch(t *testing.T) {

	expectedUrls := []string{
		"https://example.com/blog",
		"https://example.com/about",
		"https://example.com/start-outreachy",
	}

	foundUrls, err := mockSimpleFetcher.Fetch("example.com", fakeHttpClient{})

	if assert.Nil(t, err) {
		assert.Equal(t, expectedUrls, foundUrls)
	}
}

func TestGetAnchorHrefAttr(t *testing.T) {

	expectedUrl := "https://example.com/docs/"
	anchorToken := html.Token{
		Type:     html.StartTagToken,
		DataAtom: 0x1,
		Data:     "a",
		Attr:     []html.Attribute{html.Attribute{Key: "href", Val: expectedUrl}},
	}

	foundUrl, ok := mockSimpleFetcher.GetAnchorHrefAttr(anchorToken)

	if assert.True(t, ok) {
		assert.Equal(t, expectedUrl, foundUrl)
	}
}

func TestCleanUpUrl(t *testing.T) {

	parentLink := "https://example.com/doc"

	_, err := mockSimpleFetcher.CleanUpUrl("", parentLink)
	assert.NotNil(t, err, "url cleanup error : empty urls should be removed")

	_, err = mockSimpleFetcher.CleanUpUrl("/", parentLink)
	assert.NotNil(t, err, "url cleanup error : links to homepage should be removed")

	_, err = mockSimpleFetcher.CleanUpUrl("#question", parentLink)
	assert.NotNil(t, err, "url cleanup error : self-loops should be removed")

	_, err = mockSimpleFetcher.CleanUpUrl(parentLink, parentLink)
	assert.NotNil(t, err, "url cleanup error : self-loops should be removed")

	_, err = mockSimpleFetcher.CleanUpUrl("https://twitter.com", parentLink)
	assert.NotNil(t, err, "url cleanup error : external urls should be removed")

	url, err := mockSimpleFetcher.CleanUpUrl("/help", parentLink)
	if assert.Nil(t, err) {
		assert.Equal(t, url, "https://example.com/help")
	}
}
