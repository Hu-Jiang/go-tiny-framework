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
	wg sync.WaitGroup

	// visitedUrl is used to record url which has been visited.
	//
	// If your never want to crawl some pages, you can
	// init it with that urls before you execute crawl actually.
	visitedUrl sync.Map
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
		return
	}
	r.visitedUrl.Store(url, struct{}{})

	body, urls, err := fetcher.Fetch(url)
	if err != nil {
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
//
// Follow up:
// - DI, change point, interface, callback.
// - Accept interfaces, returns structs.
// - If too many args pass to function, and some are optional,
// consider functional options pattern:
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
//
// More about interface:
// 1.
// The consumer should define the interface. If you’re defining an interface
// and an implementation in the same package, you may be doing it wrong.
//
// Consumers should describe the behaviour they expect with an interface.
//
// As with all rules, there are exceptions. But the rule of thumb of accept
// interfaces return structs is a good guideline I think.
//
// The only hard and fast rule or programming that I know is that there are
// no rules, only guidelines, principals, and trade offs.
//
// Yeah, the second part “return structs/concrete types” I’m less sure about.
// But consuming interfaces that you, the caller, defines, I’m pretty sure about.
//
// 														- Dave Cheney.
//
// Ref: https://twitter.com/davecheney/status/942593128355192832
//
// 2.
// Go interfaces generally belong in the package that uses values of the interface
// type, not the package that implements those values. The implementing package
// should return concrete (usually pointer or struct) types: that way, new methods
// can be added to implementations without requiring extensive refactoring.
//
// Do not define interfaces on the implementor side of an API "for mocking";
// instead, design the API so that it can be tested using the public API of
// the real implementation.
//
// Do not define interfaces before they are used: without a realistic example of
// usage, it is too difficult to see whether an interface is even necessary, let
// alone what methods it ought to contain.
//
// Example:
//
// package consumer  // consumer.go
//
// type Thinger interface { Thing() bool }
//
// func Foo(t Thinger) string { … }
//
//
// package producer
//
// type Thinger struct{ … }
// func (t Thinger) Thing() bool { … }
//
// func NewThinger() Thinger { return Thinger{ … } }
//
//
// package consumer_test // consumer_test.go
//
// type fakeThinger struct{ … }
// func (t fakeThinger) Thing() bool { … }
// …
// if consumer.Foo(fakeThinger{…}) == "x" { … }
//
// Ref: https://github.com/golang/go/wiki/CodeReviewComments#interfaces
