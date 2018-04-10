### Usage

This concurrent web crawler outputs a site map of crawled websites.

GOPATH should be set. Refer to [this](https://golang.org/doc/code.html) for an
overview on how to organise Go code.

Download all the dependencies.

```bash
go get github.com/rs/zerolog/log
go get golang.org/x/net/html
go get github.com/stretchr/testify/assert
```
Run the program.

```bash
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

- Does not follow robots.txt, this crawler could throttle a server with too many
  connections. Ideally, there should be a delay between requests.
- Uses a global logger; should be a dependency
- url cache `collectedUrls` & url state `urlState` are global variables
- Possibly, one lock could be reduced in `web_crawler.go`
- Does not print static links.
- http client should be created once and re-used, it is safe for concurrent use
- There is no timeout on the client
