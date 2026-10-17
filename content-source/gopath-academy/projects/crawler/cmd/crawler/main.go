// Command crawler crawls a website and prints every page it finds.
//
// Usage:
//
//	crawler -url https://example.com -depth 2 -workers 4 -rate 5 -timeout 30s
//
// Only crawl sites you own or have permission to crawl.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/Priyadharshan0903/gopath-academy/projects/crawler/internal/crawl"
	"github.com/Priyadharshan0903/gopath-academy/projects/ratelimit"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("crawler", flag.ContinueOnError)
	fs.SetOutput(stderr)
	start := fs.String("url", "", "start URL (required)")
	depth := fs.Int("depth", 2, "maximum link depth")
	workers := fs.Int("workers", 4, "concurrent fetches")
	rate := fs.Float64("rate", 5, "maximum requests per second")
	maxPages := fs.Int("max", 100, "maximum pages (0 for no limit)")
	timeout := fs.Duration("timeout", 30*time.Second, "overall time limit")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *start == "" {
		fmt.Fprintln(stderr, "crawler: -url is required")
		return 2
	}
	if *rate <= 0 {
		fmt.Fprintln(stderr, "crawler: -rate must be positive")
		return 2
	}

	ctx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()

	c := &crawl.Crawler{
		Fetcher:  crawl.NewHTTPFetcher(),
		Workers:  *workers,
		MaxDepth: *depth,
		MaxPages: *maxPages,
		Limiter:  ratelimit.New(*rate, 1),
	}
	results, err := c.Crawl(ctx, *start)
	for _, r := range results {
		status := fmt.Sprint(r.Status)
		if r.Err != nil {
			status = "ERR " + r.Err.Error()
		}
		fmt.Fprintf(stdout, "%d  %-5s %s\n", r.Depth, status, r.URL)
	}
	fmt.Fprintf(stderr, "%d pages\n", len(results))
	if err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(stderr, "crawler:", err)
		return 1
	}
	return 0
}
