package crawl

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// testSite serves a small website with a cycle, a broken link, an
// external link and a slow page.
func testSite(t *testing.T, hits *atomic.Int32) *httptest.Server {
	t.Helper()
	pages := map[string]string{
		"/":         `<a href="/a">a</a> <a href="/b">b</a> <a href="https://elsewhere.example/">ext</a>`,
		"/a":        `<a href="/">home</a> <a href="/a/deep">deep</a> <a href="/missing">broken</a>`,
		"/b":        `<a href="/a#section">a again</a> <a href='/slow'>slow</a>`,
		"/a/deep":   `<a href="/a/deeper">deeper</a>`,
		"/a/deeper": `the end`,
		"/slow":     `slow page`,
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits != nil {
			hits.Add(1)
		}
		if r.URL.Path == "/slow" {
			time.Sleep(30 * time.Millisecond)
		}
		body, ok := pages[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func paths(results []Result, prefix string) []string {
	var out []string
	for _, r := range results {
		out = append(out, strings.TrimPrefix(r.URL, prefix))
	}
	slices.Sort(out)
	return out
}

func TestCrawlWholeSite(t *testing.T) {
	var hits atomic.Int32
	srv := testSite(t, &hits)
	c := &Crawler{Fetcher: NewHTTPFetcher(), Workers: 4, MaxDepth: 10}

	results, err := c.Crawl(context.Background(), srv.URL+"/")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"/", "/a", "/a/deep", "/a/deeper", "/b", "/missing", "/slow"}
	if got := paths(results, srv.URL); !slices.Equal(got, want) {
		t.Errorf("visited %v\nwant    %v", got, want)
	}
	if int(hits.Load()) != len(want) {
		t.Errorf("server got %d requests for %d pages: something was fetched twice", hits.Load(), len(want))
	}
	for _, r := range results {
		if strings.HasSuffix(r.URL, "/missing") && r.Status != http.StatusNotFound {
			t.Errorf("/missing status = %d, want 404", r.Status)
		}
	}
}

func TestMaxDepth(t *testing.T) {
	srv := testSite(t, nil)
	c := &Crawler{Fetcher: NewHTTPFetcher(), Workers: 2, MaxDepth: 1}
	results, err := c.Crawl(context.Background(), srv.URL+"/")
	if err != nil {
		t.Fatal(err)
	}
	if got := paths(results, srv.URL); !slices.Equal(got, []string{"/", "/a", "/b"}) {
		t.Errorf("depth 1 visited %v", got)
	}
}

func TestMaxPages(t *testing.T) {
	srv := testSite(t, nil)
	c := &Crawler{Fetcher: NewHTTPFetcher(), Workers: 3, MaxDepth: 10, MaxPages: 3}
	results, err := c.Crawl(context.Background(), srv.URL+"/")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Errorf("got %d results, want exactly 3", len(results))
	}
}

func TestTimeout(t *testing.T) {
	srv := testSite(t, nil)
	c := &Crawler{Fetcher: NewHTTPFetcher(), Workers: 1, MaxDepth: 10, Limiter: slowLimiter{50 * time.Millisecond}}
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	_, err := c.Crawl(ctx, srv.URL+"/")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want DeadlineExceeded", err)
	}
}

type slowLimiter struct{ d time.Duration }

func (s slowLimiter) Wait(ctx context.Context) error {
	select {
	case <-time.After(s.d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestNoGoroutineLeaks(t *testing.T) {
	srv := testSite(t, nil)
	before := runtime.NumGoroutine()
	for range 5 {
		c := &Crawler{Fetcher: NewHTTPFetcher(), Workers: 8, MaxDepth: 10}
		if _, err := c.Crawl(context.Background(), srv.URL+"/"); err != nil {
			t.Fatal(err)
		}
	}
	// Idle HTTP keep-alive connections own goroutines too; close them.
	http.DefaultClient.CloseIdleConnections()
	deadline := time.Now().Add(2 * time.Second)
	for runtime.NumGoroutine() > before+2 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > before+2 {
		t.Errorf("%d goroutines still running after crawls finished (before: %d)", n, before)
	}
}
