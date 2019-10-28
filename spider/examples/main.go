package main

import (
	"fmt"
	"os"
	"spider"
	"spider/links"
)

type FetcherFunc func(url string) (body string, urls []string, err error)

// Fetch warp f to satisfy spider.Fetcher interface.
func (f FetcherFunc) Fetch(url string) (body string, urls []string, err error) {
	return f(url)
}

// fetch warp Extract function to match FetcherFunc.
func fetch(url string) (body string, urls []string, err error) {
	list, err := links.Extract(url)
	if err != nil {
		return "", nil, err
	}
	return "", list, nil
}

func main() {
	spider.DistributedCrawl("https://www.hao123.com", 2, FetcherFunc(fetch), os.Stdout)
	fmt.Println("-------------------------------------------------------------------")
	spider.SequentialCrawl("https://www.hao123.com", 2, FetcherFunc(fetch), os.Stdout)
}
