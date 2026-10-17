// Package crawl implements a concurrent, polite web crawler.
//
// One coordinator goroutine owns all crawl state (the queue and the set
// of seen URLs), and a fixed pool of workers does the fetching. Workers
// never touch shared state, so the only synchronization is channels.
package crawl

import (
	"context"
	"net/url"
)

// Limiter delays work to respect a rate limit. *ratelimit.Bucket from
// the ratelimit project satisfies it.
type Limiter interface {
	Wait(ctx context.Context) error
}

// Result describes one crawled URL.
type Result struct {
	URL    string
	Depth  int
	Status int
	Err    error
}

// Crawler crawls pages on the same host as the start URL.
type Crawler struct {
	Fetcher  Fetcher
	Workers  int     // concurrent fetches; values below 1 mean 1
	MaxDepth int     // 0 crawls only the start page
	MaxPages int     // stop after this many pages; 0 means no limit
	Limiter  Limiter // optional
}

type task struct {
	url   string
	depth int
}

type outcome struct {
	task
	page Page
	err  error
}

// Crawl visits start and the pages it links to, breadth first, until
// MaxDepth or MaxPages is reached or ctx is done. It returns one Result
// per visited URL, in the order they finished.
func (c *Crawler) Crawl(ctx context.Context, start string) ([]Result, error) {
	startURL, err := url.Parse(start)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // stops the workers when Crawl returns

	tasks := make(chan task)
	outcomes := make(chan outcome)

	workers := max(1, c.Workers)
	for range workers {
		go c.worker(ctx, tasks, outcomes)
	}

	// Coordinator state: only this goroutine reads or writes it.
	seen := map[string]bool{normalize(startURL): true}
	queue := []task{{url: startURL.String(), depth: 0}}
	inFlight := 0
	var results []Result

	for len(queue) > 0 || inFlight > 0 {
		// The nil-channel trick (lesson 6.3): only enable the send case
		// when there is something to send.
		var send chan<- task
		var next task
		if len(queue) > 0 && (c.MaxPages == 0 || len(results)+inFlight < c.MaxPages) {
			send, next = tasks, queue[0]
		}

		select {
		case send <- next:
			queue = queue[1:]
			inFlight++
		case o := <-outcomes:
			inFlight--
			results = append(results, Result{URL: o.url, Depth: o.depth, Status: o.page.Status, Err: o.err})
			if o.err != nil || o.depth >= c.MaxDepth {
				continue
			}
			for _, link := range o.page.Links {
				u, err := url.Parse(link)
				if err != nil || u.Host != startURL.Host || (u.Scheme != "http" && u.Scheme != "https") {
					continue // stay on the same site
				}
				key := normalize(u)
				if !seen[key] {
					seen[key] = true
					queue = append(queue, task{url: u.String(), depth: o.depth + 1})
				}
			}
		case <-ctx.Done():
			return results, ctx.Err()
		}
		if c.MaxPages > 0 && len(results) >= c.MaxPages && inFlight == 0 {
			break
		}
	}
	return results, nil
}

func (c *Crawler) worker(ctx context.Context, tasks <-chan task, outcomes chan<- outcome) {
	for {
		var t task
		select {
		case t = <-tasks:
		case <-ctx.Done():
			return
		}
		o := outcome{task: t}
		if c.Limiter != nil {
			o.err = c.Limiter.Wait(ctx)
		}
		if o.err == nil {
			o.page, o.err = c.Fetcher.Fetch(ctx, t.url)
		}
		select {
		case outcomes <- o:
		case <-ctx.Done():
			return
		}
	}
}

// normalize removes parts of a URL that do not identify a different page.
func normalize(u *url.URL) string {
	n := *u
	n.Fragment = ""
	if n.Path == "" {
		n.Path = "/"
	}
	return n.String()
}
