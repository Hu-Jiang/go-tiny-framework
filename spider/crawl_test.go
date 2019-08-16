package spider_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"sort"
	"spider"
	"strings"
	"testing"
	"time"
)

func TestSequentialCrawl(t *testing.T) {
	var buf bytes.Buffer
	spider.SequentialCrawl("https://golang.org/", 4, fetcher, &buf)

	assertResults(t, gotResults(&buf), wantResults())

}
func TestDistributedCrawl(t *testing.T) {
	var buf bytes.Buffer
	spider.DistributedCrawl("https://golang.org/", 4, fetcher, &buf)
	assertResults(t, gotResults(&buf), wantResults())
}

func BenchmarkSequentialCrawl(b *testing.B) {
	for i := 0; i < b.N; i++ {
		spider.SequentialCrawl("https://golang.org/", 4, fetcher, ioutil.Discard)
	}
}
func BenchmarkDistributedCrawl(b *testing.B) {
	for i := 0; i < b.N; i++ {
		spider.DistributedCrawl("https://golang.org/", 4, fetcher, ioutil.Discard)
	}
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	time.Sleep(1 * time.Millisecond) // simulate html parse time
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}

func wantResults() []string {
	var results []string
	for url, res := range fetcher {
		results = append(results, fmt.Sprintf("found: %s %q", url, res.body))
	}

	return results
}

func gotResults(buf *bytes.Buffer) []string {
	strs := strings.Trim(buf.String(), "\n")
	got := strings.Split(strs, "\n")
	return got
}

func assertResults(t *testing.T, got, want []string) {
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
