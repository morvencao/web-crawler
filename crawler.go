package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type Cache struct {
	visited map[string]bool
	mux sync.Mutex
}

type Response struct {
	url string
	body string
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, ch chan Response, cache Cache) {
	defer close(ch)
	if depth <= 0 {
		return
	}

	cache.mux.Lock()
	if cache.visited[url] {
		cache.mux.Unlock()
		return
	}
	cache.visited[url] = true
	cache.mux.Unlock()

	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}

	ch <- Response{url, body}
	chs := make([]chan Response, len(urls))
	for i, u := range urls {
		chs[i] = make(chan Response)
		go Crawl(u, depth-1, fetcher, chs[i], cache)
	}

	for i := range chs {
		for resp := range chs[i] {
			ch <- resp
		}
	}
	return
}

func main() {
	var ch = make(chan Response)
	go Crawl("https://golang.org/", 4, fetcher, ch, Cache{visited: make(map[string]bool)})
	for resp := range ch {
		fmt.Printf("found: %s %q\n", resp.body, resp.url)
	}
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
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

