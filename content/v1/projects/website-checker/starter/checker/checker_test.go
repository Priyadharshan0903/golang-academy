package checker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response(r *http.Request, code int, body io.ReadCloser) *http.Response {
	return &http.Response{StatusCode: code, Header: make(http.Header), Body: body, Request: r}
}
func engine(t *testing.T, workers int, timeout time.Duration, transport transportFunc) *Checker {
	t.Helper()
	c, err := New(Config{Workers: workers, RequestTimeout: timeout, Client: &http.Client{Transport: transport}})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

type trackedBody struct {
	io.Reader
	closes *atomic.Int64
}

func (b trackedBody) Close() error { b.closes.Add(1); return nil }

func TestConfiguration(t *testing.T) {
	for _, config := range []Config{{Workers: 0, RequestTimeout: time.Second}, {Workers: 9, RequestTimeout: time.Second}, {Workers: 1}} {
		if _, err := New(config); err == nil {
			t.Fatalf("accepted invalid config: %+v", config)
		}
	}
}

func TestResultsStatusSortingDuplicatesAndCleanup(t *testing.T) {
	var closed atomic.Int64
	c := engine(t, 3, time.Second, func(r *http.Request) (*http.Response, error) {
		status := 200
		if r.URL.Path == "/bad" {
			status = 503
		}
		return response(r, status, trackedBody{Reader: strings.NewReader("ok"), closes: &closed}), nil
	})
	urls := []string{"https://example.test/z", "https://example.test/bad", "https://example.test/a", "https://example.test/a"}
	got, err := c.Check(context.Background(), urls)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"https://example.test/a", "https://example.test/a", "https://example.test/bad", "https://example.test/z"}
	if len(got) != len(want) {
		t.Fatalf("results: %d; want %d", len(got), len(want))
	}
	for i, result := range got {
		if result.URL != want[i] {
			t.Fatalf("position %d: %q; want %q", i, result.URL, want[i])
		}
	}
	if got[2].StatusCode != 503 || got[2].Error != "" {
		t.Fatalf("HTTP failure must remain a status: %+v", got[2])
	}
	if closed.Load() != int64(len(urls)) {
		t.Fatalf("closed %d bodies; want %d", closed.Load(), len(urls))
	}
}

func TestWorkerBound(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var active, maximum atomic.Int64
	entered := make(chan struct{}, 32)
	release := make(chan struct{})
	defer close(release)
	c := engine(t, 8, time.Second, func(r *http.Request) (*http.Response, error) {
		current := active.Add(1)
		defer active.Add(-1)
		for {
			old := maximum.Load()
			if current <= old || maximum.CompareAndSwap(old, current) {
				break
			}
		}
		entered <- struct{}{}
		select {
		case <-release:
		case <-r.Context().Done():
			return nil, r.Context().Err()
		}
		return response(r, 200, io.NopCloser(strings.NewReader("ok"))), nil
	})
	urls := make([]string, 24)
	for i := range urls {
		urls[i] = fmt.Sprintf("https://example.test/%02d", i)
	}
	type outcome struct {
		results []Result
		err     error
	}
	done := make(chan outcome, 1)
	go func() { results, err := c.Check(ctx, urls); done <- outcome{results, err} }()
	for i := 0; i < 8; i++ {
		select {
		case <-entered:
		case <-ctx.Done():
			t.Fatal("workers did not start:", ctx.Err())
		}
	}
	if active.Load() != 8 {
		t.Fatalf("active = %d; want 8", active.Load())
	}
	// Release all workers without relying on a sleep or scheduling order.
	for i := 0; i < 24; i++ {
		select {
		case release <- struct{}{}:
		case <-ctx.Done():
			t.Fatal("workers did not finish")
		}
	}
	select {
	case outcome := <-done:
		if outcome.err != nil || len(outcome.results) != 24 {
			t.Fatalf("results=%d error=%v", len(outcome.results), outcome.err)
		}
	case <-ctx.Done():
		t.Fatal("run did not complete")
	}
	if maximum.Load() > 8 {
		t.Fatalf("worker limit exceeded: %d", maximum.Load())
	}
	if active.Load() != 0 {
		t.Fatal("request remained active after Check returned")
	}
}

func TestCancellationStopsInFlightWork(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{}, 8)
	var active atomic.Int64
	c := engine(t, 4, time.Minute, func(r *http.Request) (*http.Response, error) {
		active.Add(1)
		defer active.Add(-1)
		started <- struct{}{}
		<-r.Context().Done()
		return nil, r.Context().Err()
	})
	done := make(chan error, 1)
	go func() {
		_, err := c.Check(ctx, []string{"https://example.test/a", "https://example.test/b", "https://example.test/c", "https://example.test/d", "https://example.test/e"})
		done <- err
	}()
	for i := 0; i < 4; i++ {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatal("request did not start")
		}
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error=%v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("canceled run did not finish")
	}
	if active.Load() != 0 {
		t.Fatal("canceled requests still active")
	}
}

func TestPerRequestDeadlineIsAResult(t *testing.T) {
	c := engine(t, 1, 10*time.Millisecond, func(r *http.Request) (*http.Response, error) { <-r.Context().Done(); return nil, r.Context().Err() })
	results, err := c.Check(context.Background(), []string{"https://example.test/slow"})
	if err != nil || len(results) != 1 || !strings.Contains(results[0].Error, "deadline exceeded") {
		t.Fatalf("results=%+v error=%v", results, err)
	}
}

func TestValidateEverythingBeforeNetworkWork(t *testing.T) {
	var requests atomic.Int64
	c := engine(t, 2, time.Second, func(r *http.Request) (*http.Response, error) {
		requests.Add(1)
		return response(r, 200, io.NopCloser(strings.NewReader(""))), nil
	})
	for _, bad := range []string{"", "relative/path", "ftp://example.test/a", "https://user:secret@example.test", "https://example.test/#fragment"} {
		if _, err := c.Check(context.Background(), []string{"https://example.test/valid", bad}); err == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
	if requests.Load() != 0 {
		t.Fatalf("started %d requests for invalid input", requests.Load())
	}
}

func TestEmptyAndAlreadyCanceled(t *testing.T) {
	c := engine(t, 1, time.Second, func(*http.Request) (*http.Response, error) {
		t.Error("unexpected request")
		return nil, errors.New("unexpected")
	})
	if got, err := c.Check(context.Background(), nil); err != nil || len(got) != 0 {
		t.Fatalf("got=%v err=%v", got, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Check(ctx, []string{"https://example.test"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
}

type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, errors.New("broken stream") }
func TestBodyReadErrorStillCloses(t *testing.T) {
	var closed atomic.Int64
	c := engine(t, 1, time.Second, func(r *http.Request) (*http.Response, error) {
		return response(r, 200, trackedBody{Reader: brokenReader{}, closes: &closed}), nil
	})
	got, err := c.Check(context.Background(), []string{"https://example.test"})
	if err != nil || len(got) != 1 || !strings.Contains(got[0].Error, "broken stream") || closed.Load() != 1 {
		t.Fatalf("got=%v err=%v closed=%d", got, err, closed.Load())
	}
}

func TestTransportFailureIsAResult(t *testing.T) {
	c := engine(t, 1, time.Second, func(*http.Request) (*http.Response, error) { return nil, errors.New("network unavailable") })
	got, err := c.Check(context.Background(), []string{"https://example.test"})
	if err != nil || len(got) != 1 || !strings.Contains(got[0].Error, "network unavailable") {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestURLLimit(t *testing.T) {
	c := engine(t, 1, time.Second, func(*http.Request) (*http.Response, error) {
		t.Error("unexpected request")
		return nil, errors.New("unexpected")
	})
	if _, err := c.Check(context.Background(), make([]string, MaxURLs+1)); err == nil {
		t.Fatal("accepted too many URLs")
	}
}
