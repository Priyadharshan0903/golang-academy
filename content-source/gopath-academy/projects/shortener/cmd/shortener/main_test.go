package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/pgtest"
)

// TestRun starts the real program on free ports, uses it over HTTP,
// and shuts it down by cancelling the context, like a SIGTERM would.
func TestRun(t *testing.T) {
	testRun(t, map[string]string{})
}

// TestRunWithPostgresAndRedis runs the same test with real storage.
func TestRunWithPostgresAndRedis(t *testing.T) {
	if os.Getenv("TEST_REDIS_ADDR") == "" {
		t.Skip("TEST_REDIS_ADDR not set")
	}
	testRun(t, map[string]string{
		"DATABASE_URL": pgtest.URL(t),
		"REDIS_ADDR":   os.Getenv("TEST_REDIS_ADDR"),
	})
}

func testRun(t *testing.T, env map[string]string) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	env["ADDR"] = "127.0.0.1:0"
	env["ADMIN_ADDR"] = "127.0.0.1:0"
	env["BASE_URL"] = "https://sho.rt"
	pr, pw := io.Pipe()
	done := make(chan error, 1)
	go func() {
		done <- run(ctx, func(k string) string { return env[k] }, pw)
		pw.Close()
	}()

	// Read the log until both servers report their addresses.
	addrs := map[string]string{}
	logs := bufio.NewScanner(pr)
	for len(addrs) < 2 && logs.Scan() {
		var line struct{ Msg, Server, Addr string }
		json.Unmarshal(logs.Bytes(), &line)
		if line.Msg == "listening" {
			addrs[line.Server] = "http://" + line.Addr
		}
	}
	go io.Copy(io.Discard, pr) // keep draining the log

	client := &http.Client{
		Timeout:       5 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	waitReady(t, client, addrs["api"]+"/readyz")

	resp, err := client.Post(addrs["api"]+"/api/links", "application/json",
		strings.NewReader(`{"url":"https://go.dev","code":"go-home"}`))
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: %v %v", resp.Status, err)
	}
	resp.Body.Close()

	resp, err = client.Get(addrs["api"] + "/go-home")
	if err != nil || resp.StatusCode != http.StatusFound || resp.Header.Get("Location") != "https://go.dev" {
		t.Fatalf("redirect: %v %v", resp.Status, err)
	}
	resp.Body.Close()

	resp, err = client.Get(addrs["admin"] + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(body), `shortener_http_requests_total{code="302",route="GET /{code}"} 1`) ||
		!strings.Contains(string(body), "go_goroutines") {
		t.Errorf("metrics incomplete:\n%s", body)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("run returned %v, want nil after a clean shutdown", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("run did not return after cancel")
	}
}

func waitReady(t *testing.T, c *http.Client, url string) {
	t.Helper()
	for range 50 {
		resp, err := c.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("%s never became ready", url)
}

func TestRunBadConfig(t *testing.T) {
	err := run(t.Context(), func(k string) string {
		return map[string]string{"RATE_LIMIT": "fast"}[k]
	}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "RATE_LIMIT") {
		t.Fatalf("err = %v", err)
	}
}
