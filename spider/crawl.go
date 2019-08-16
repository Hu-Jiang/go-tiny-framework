package spider

import (
	"fmt"
	"io"
	"sync"
)

// NOTE:
// We can use callback to substitute Fetcher interface, such as:
// type Fetcher func(url string) (body string, urls []string, err error)
type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// A record represents some record when crawl, such as:
// url which has been crawl and
// goroutine when spider is Distributed.
type Record struct {
	visitedUrl sync.Map // visitedUrl is url which has been visited
	wg         sync.WaitGroup
}

// Sequential crawl url sequentially, waiting for each crawl task to
// complete before crawl the next url.
func SequentialCrawl(url string, depth int, fetcher Fetcher, w io.Writer) {
	r := &Record{}
	r.SequentialCrawl(url, depth, fetcher, w)
}

func (r *Record) SequentialCrawl(url string, depth int, fetcher Fetcher, w io.Writer) {
	if depth <= 0 {
		return
	}

	if _, ok := r.visitedUrl.Load(url); ok {
		// fmt.Printf("has been visited: %s\n", url)
		return
	}
	r.visitedUrl.Store(url, struct{}{})

	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		// fmt.Println(err)
		return
	}
	fmt.Fprintf(w, "found: %s %q\n", url, body)
	for _, u := range urls {
		r.SequentialCrawl(u, depth-1, fetcher, w)
	}
	return
}

// DistributedCrawl uses fetcher to recursively crawl
// pages starting with url distributed, to a maximum of depth.
func DistributedCrawl(url string, depth int, fetcher Fetcher, w io.Writer) {
	r := &Record{}
	r.wg.Add(1)
	r.DistributedCrawl(url, depth, fetcher, w)
	r.wg.Wait()
}

func (r *Record) DistributedCrawl(url string, depth int, fetcher Fetcher, w io.Writer) {
	defer r.wg.Done()
	if depth <= 0 {
		return
	}
	if _, ok := r.visitedUrl.Load(url); ok {
		return
	}
	r.visitedUrl.Store(url, struct{}{})

	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		// fmt.Println(err)
		return
	}
	fmt.Fprintf(w, "found: %s %q\n", url, body)
	for _, u := range urls {
		r.wg.Add(1)
		go r.DistributedCrawl(u, depth-1, fetcher, w)
	}
	return
}

// Ref:
// https://tour.golang.org/concurrency/10

// Follow up:
// - DI, change point, interface, callback.
// - Pass args by interface, return struct.
// - If to many args pass to function, and some are optional,
// consider functional options pattern:
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
