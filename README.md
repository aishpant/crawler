### Usage

This web crawler outputs a site map of crawled websites.

GOPATH should be set.

First, download all the dependencies.

```bash
go get github.com/rs/zerolog/log
go get golang.org/x/net/html
go get github.com/stretchr/testify/assert
```

Then, copy the source files and run the program. Refer to
[this](https://golang.org/doc/code.html) for an overview on how to organise Go
code.

```bash
mkdir $GOPATH/src/../crawler
# copy the source files to the crawler directory
go build
./crawler -url https://golang.org/ -depth 4 -output out.txt
# or run the next command to suppress the logs
./crawler -url https://golang.org/ -depth 4 -output out.txt &> /dev/null
```

The output sitemap is written to `out.txt`.

This web crawler is inspired by the Web Crawler exercise from the
[gotour](https://tour.golang.org/concurrency/10)


Run `go test` command to run all the tests.

### Issues

- Does not follow robots.txt
- Uses a global logger; should be a dependency
- url cache `collectedUrls` & url state `urlState` are again stored in global variables
- Possibly, one lock could be reduced in `web_crawler.go`
