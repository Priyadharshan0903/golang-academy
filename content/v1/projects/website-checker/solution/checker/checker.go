// Package checker performs bounded, cancelable HTTP health checks.
package checker

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"
)

const MaxWorkers = 8
const MaxURLs = 1000

type Config struct {
	Workers        int
	RequestTimeout time.Duration
	Client         *http.Client
}

type Result struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	Error      string `json:"error,omitempty"`
}

// Checker may be shared by callers. Do not mutate its client's configuration
// during a run; the client's transport must support concurrent use.
type Checker struct {
	workers int
	timeout time.Duration
	client  *http.Client
}

func New(config Config) (*Checker, error) {
	if config.Workers < 1 || config.Workers > MaxWorkers {
		return nil, fmt.Errorf("workers must be between 1 and %d", MaxWorkers)
	}
	if config.RequestTimeout <= 0 {
		return nil, fmt.Errorf("request timeout must be positive")
	}
	client := config.Client
	if client == nil {
		client = &http.Client{}
	}
	return &Checker{workers: config.Workers, timeout: config.RequestTimeout, client: client}, nil
}

// Check validates the entire input before making requests. With no run-level
// cancellation it returns one result per URL, including duplicates and failures.
// On cancellation it returns completed results plus ctx.Err(). Results are sorted
// by URL. Individual request failures belong to Result, not the returned error.
func (c *Checker) Check(ctx context.Context, urls []string) ([]Result, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context must not be nil")
	}
	if len(urls) > MaxURLs {
		return nil, fmt.Errorf("at most %d URLs are supported", MaxURLs)
	}
	for i, raw := range urls {
		parsed, err := url.Parse(raw)
		if err != nil || parsed.Hostname() == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.User != nil || parsed.Fragment != "" {
			return nil, fmt.Errorf("URL %d must be an absolute HTTP(S) URL without credentials or a fragment", i+1)
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	results := make([]Result, 0, len(urls))
	if len(urls) == 0 {
		return results, nil
	}
	jobs := make(chan string)
	completed := make(chan Result)
	var workers sync.WaitGroup
	count := min(c.workers, len(urls))
	for i := 0; i < count; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case raw, ok := <-jobs:
					if !ok {
						return
					}
					if ctx.Err() != nil {
						return
					}
					result := c.checkOne(ctx, raw)
					// The collector always drains this channel, even after cancellation.
					// This send therefore cannot strand a worker when the run is stopped.
					completed <- result
				}
			}
		}()
	}
	// A single producer owns closing jobs. It observes cancellation while sending.
	go func() {
		defer close(jobs)
		for _, raw := range urls {
			select {
			case <-ctx.Done():
				return
			case jobs <- raw:
			}
		}
	}()
	// Only this coordinator closes the result channel, after all senders exit.
	go func() { workers.Wait(); close(completed) }()
	for result := range completed {
		results = append(results, result)
	}
	sort.SliceStable(results, func(i, j int) bool { return results[i].URL < results[j].URL })
	return results, ctx.Err()
}

func (c *Checker) checkOne(parent context.Context, raw string) Result {
	result := Result{URL: raw}
	ctx, cancel := context.WithTimeout(parent, c.timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	response, err := c.client.Do(request)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer response.Body.Close()
	result.StatusCode = response.StatusCode
	// Discard at most 4 KiB. A large or endless response must not dominate a run.
	// A small body read to EOF may allow connection reuse. A read failure remains
	// observable; a limit by itself does not claim the whole body was consumed.
	if _, err := io.Copy(io.Discard, io.LimitReader(response.Body, 4096)); err != nil {
		result.Error = fmt.Sprintf("read response: %v", err)
	}
	return result
}
