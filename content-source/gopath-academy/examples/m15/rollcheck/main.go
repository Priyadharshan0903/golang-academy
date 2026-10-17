// Command rollcheck sends a steady stream of requests to a URL and
// counts the results, to check that a rolling update loses no requests
// (lesson 15.2):
//
//	rollcheck -url http://shortener.localtest.me/healthz -duration 60s &
//	kubectl -n shortener rollout restart deployment/shortener
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync"
	"time"
)

func main() {
	url := flag.String("url", "", "URL to request (required)")
	host := flag.String("host", "", "Host header to send, for ingress routing")
	duration := flag.Duration("duration", 30*time.Second, "how long to run")
	workers := flag.Int("workers", 4, "concurrent clients")
	flag.Parse()
	if *url == "" {
		flag.Usage()
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, *duration)
	defer cancel()

	r := run(ctx, *url, *host, *workers)
	fmt.Print(r)
	if r.failed() > 0 {
		os.Exit(1)
	}
}

// [start run]

// result counts responses by status code, and transport errors by message.
type result struct {
	mu       sync.Mutex
	statuses map[int]int
	errors   map[string]int
	total    int
}

func (r *result) failed() int {
	n := 0
	for code, c := range r.statuses {
		if code >= 500 {
			n += c
		}
	}
	for _, c := range r.errors {
		n += c
	}
	return n
}

func (r *result) String() string {
	s := fmt.Sprintf("%d requests, %d failed\n", r.total, r.failed())
	codes := make([]int, 0, len(r.statuses))
	for c := range r.statuses {
		codes = append(codes, c)
	}
	sort.Ints(codes)
	for _, c := range codes {
		s += fmt.Sprintf("  HTTP %d: %d\n", c, r.statuses[c])
	}
	for e, c := range r.errors {
		s += fmt.Sprintf("  error %q: %d\n", e, c)
	}
	return s
}

func run(ctx context.Context, url, host string, workers int) *result {
	r := &result{statuses: map[int]int{}, errors: map[string]int{}}
	// A new connection for every request, like many independent clients:
	// keep-alive connections would hide what happens to new connections.
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: &http.Transport{DisableKeepAlives: true},
	}
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ctx.Err() == nil {
				req, _ := http.NewRequest(http.MethodGet, url, nil)
				if host != "" {
					req.Host = host
				}
				resp, err := client.Do(req)
				r.mu.Lock()
				r.total++
				if err != nil {
					r.errors[err.Error()]++
				} else {
					r.statuses[resp.StatusCode]++
				}
				r.mu.Unlock()
				if err == nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				}
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}
	wg.Wait()
	return r
}

// [end run]
