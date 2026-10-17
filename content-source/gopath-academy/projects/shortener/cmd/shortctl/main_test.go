package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/httpapi"
	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/link"
)

// newServer runs the real API handler with an in-memory store.
func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(nil)
	base, _ := url.Parse(srv.URL)
	srv.Config.Handler = httpapi.New(httpapi.Options{
		Service: link.NewService(link.NewMemoryStore()),
		BaseURL: base,
		Logger:  slog.New(slog.DiscardHandler),
	})
	t.Cleanup(srv.Close)
	return srv
}

type result struct {
	code           int
	stdout, stderr string
}

func shortctl(t *testing.T, vars map[string]string, args ...string) result {
	t.Helper()
	e := env{
		getenv:    func(k string) string { return vars[k] },
		readFile:  func(string) ([]byte, error) { return nil, fs.ErrNotExist },
		configDir: func() (string, error) { return "/nonexistent", nil },
	}
	var out, errOut bytes.Buffer
	code := run(context.Background(), args, e, &out, &errOut)
	return result{code, out.String(), errOut.String()}
}

func TestCreateAndGet(t *testing.T) {
	srv := newServer(t)
	vars := map[string]string{"SHORTCTL_SERVER": srv.URL}

	r := shortctl(t, vars, "create", "-code", "godoc", "https://go.dev/doc")
	if r.code != 0 || r.stdout != srv.URL+"/godoc\n" {
		t.Fatalf("create: %+v", r)
	}
	r = shortctl(t, vars, "get", "godoc")
	if r.code != 0 || !strings.HasPrefix(r.stdout, srv.URL+"/godoc -> https://go.dev/doc (created ") {
		t.Fatalf("get: %+v", r)
	}
	r = shortctl(t, vars, "-output", "json", "get", "godoc")
	var l apiLink
	if err := json.Unmarshal([]byte(r.stdout), &l); err != nil || l.URL != "https://go.dev/doc" {
		t.Fatalf("get -output json: %+v (%v)", r, err)
	}
}

func TestDelete(t *testing.T) {
	srv := newServer(t)
	vars := map[string]string{"SHORTCTL_SERVER": srv.URL}
	shortctl(t, vars, "create", "-code", "temp", "https://go.dev")
	if r := shortctl(t, vars, "delete", "temp"); r.code != 0 || r.stdout != "deleted temp\n" {
		t.Fatalf("delete: %+v", r)
	}
	if r := shortctl(t, vars, "delete", "temp"); r.code != 1 || !strings.Contains(r.stderr, "HTTP 404") {
		t.Fatalf("second delete: %+v", r)
	}
}

func TestErrors(t *testing.T) {
	srv := newServer(t)
	vars := map[string]string{"SHORTCTL_SERVER": srv.URL}
	shortctl(t, vars, "create", "-code", "taken", "https://go.dev")

	tests := []struct {
		name   string
		args   []string
		code   int
		stderr string
	}{
		{"conflict", []string{"create", "-code", "taken", "https://go.dev"}, 1, `create "taken": code already in use (HTTP 409)`},
		{"not found", []string{"get", "nothere"}, 1, "link not found (HTTP 404)"},
		{"invalid url", []string{"create", "ftp://x"}, 1, "HTTP 422"},
		{"no command", nil, 2, "usage: shortctl"},
		{"unknown command", []string{"remove", "x"}, 2, `unknown command "remove"`},
		{"missing url", []string{"create"}, 2, "exactly one URL"},
		{"bad output", []string{"-output", "yaml", "get", "x"}, 2, "text or json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := shortctl(t, vars, tt.args...)
			if r.code != tt.code || !strings.Contains(r.stderr, tt.stderr) {
				t.Errorf("exit %d, stderr %q; want exit %d and %q", r.code, r.stderr, tt.code, tt.stderr)
			}
		})
	}
}

func TestServerDown(t *testing.T) {
	srv := newServer(t)
	srv.Close()
	r := shortctl(t, map[string]string{"SHORTCTL_SERVER": srv.URL}, "get", "abc")
	if r.code != 1 || !strings.Contains(r.stderr, "cannot reach") {
		t.Fatalf("%+v", r)
	}
}
