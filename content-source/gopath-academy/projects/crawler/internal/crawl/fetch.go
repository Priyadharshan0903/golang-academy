package crawl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

// Page is what a Fetcher returns for one URL.
type Page struct {
	Status int
	Links  []string // absolute URLs found on the page
}

// Fetcher downloads a page and extracts its links. The crawler only
// depends on this interface, so tests can use a fake (lesson 4.2).
type Fetcher interface {
	Fetch(ctx context.Context, rawURL string) (Page, error)
}

// HTTPFetcher fetches pages over HTTP.
type HTTPFetcher struct {
	Client  *http.Client
	MaxBody int64 // bytes to read per page; 0 means 1 MiB
}

// NewHTTPFetcher returns a fetcher with a sensible client timeout.
func NewHTTPFetcher() *HTTPFetcher {
	return &HTTPFetcher{Client: &http.Client{Timeout: 10 * time.Second}}
}

// hrefPattern finds href attributes. A regular expression is NOT a
// real HTML parser; production crawlers should use golang.org/x/net/html.
// It keeps this project dependency-free and is good enough for the tests.
var hrefPattern = regexp.MustCompile(`(?i)href\s*=\s*["']([^"'#]+)`)

// Fetch downloads rawURL and returns its status and links.
func (f *HTTPFetcher) Fetch(ctx context.Context, rawURL string) (Page, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return Page{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "gopath-academy-crawler/1.0")

	resp, err := f.Client.Do(req)
	if err != nil {
		return Page{}, err
	}
	defer resp.Body.Close()

	limit := f.MaxBody
	if limit == 0 {
		limit = 1 << 20
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit))
	if err != nil {
		return Page{Status: resp.StatusCode}, fmt.Errorf("read body: %w", err)
	}

	base := resp.Request.URL // after redirects
	var links []string
	for _, m := range hrefPattern.FindAllSubmatch(body, -1) {
		ref, err := url.Parse(string(m[1]))
		if err != nil {
			continue
		}
		links = append(links, base.ResolveReference(ref).String())
	}
	return Page{Status: resp.StatusCode, Links: links}, nil
}
