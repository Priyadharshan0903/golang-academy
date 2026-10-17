package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/link"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type env struct {
	h    http.Handler
	logs *bytes.Buffer
	reg  *prometheus.Registry
}

func newEnv(t *testing.T, store link.Store, limiter *ClientLimiter) env {
	t.Helper()
	base, _ := url.Parse("https://sho.rt")
	logs := &bytes.Buffer{}
	reg := prometheus.NewRegistry()
	var ready atomic.Bool
	ready.Store(true)
	h := New(Options{
		Service: link.NewService(store),
		BaseURL: base,
		Logger:  slog.New(slog.NewJSONHandler(logs, nil)),
		Ready:   &ready,
		Metrics: NewMetrics(reg),
		Limiter: limiter,
	})
	return env{h, logs, reg}
}

func (e env) do(method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, req)
	return rec
}

func TestCreateResolveRedirect(t *testing.T) {
	e := newEnv(t, link.NewMemoryStore(), nil)

	rec := e.do("POST", "/api/links", `{"url":"https://go.dev/doc","code":"godoc"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rec.Code, rec.Body)
	}
	if loc := rec.Header().Get("Location"); loc != "/api/links/godoc" {
		t.Errorf("Location = %q", loc)
	}
	var got linkResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Code != "godoc" || got.URL != "https://go.dev/doc" || got.ShortURL != "https://sho.rt/godoc" || got.CreatedAt.IsZero() {
		t.Errorf("response = %+v", got)
	}

	rec = e.do("GET", "/api/links/godoc", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"short_url":"https://sho.rt/godoc"`) {
		t.Errorf("get: %d %s", rec.Code, rec.Body)
	}

	rec = e.do("GET", "/godoc", "")
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "https://go.dev/doc" {
		t.Errorf("redirect: %d Location=%q", rec.Code, rec.Header().Get("Location"))
	}
}

func TestGeneratedCode(t *testing.T) {
	e := newEnv(t, link.NewMemoryStore(), nil)
	rec := e.do("POST", "/api/links", `{"url":"https://example.com"}`)
	var got linkResponse
	json.Unmarshal(rec.Body.Bytes(), &got)
	if rec.Code != http.StatusCreated || len(got.Code) != 7 {
		t.Fatalf("%d %+v", rec.Code, got)
	}
}

func TestErrorMapping(t *testing.T) {
	e := newEnv(t, link.NewMemoryStore(), nil)
	e.do("POST", "/api/links", `{"url":"https://go.dev","code":"taken"}`)

	tests := []struct {
		name, method, path, body string
		want                     int
	}{
		{"code in use", "POST", "/api/links", `{"url":"https://go.dev","code":"taken"}`, http.StatusConflict},
		{"bad scheme", "POST", "/api/links", `{"url":"ftp://go.dev"}`, http.StatusUnprocessableEntity},
		{"bad code", "POST", "/api/links", `{"url":"https://go.dev","code":"a b"}`, http.StatusUnprocessableEntity},
		{"unknown field", "POST", "/api/links", `{"link":"https://go.dev"}`, http.StatusBadRequest},
		{"not json", "POST", "/api/links", `hello`, http.StatusBadRequest},
		{"no content type", "POST", "/api/links", ``, http.StatusUnsupportedMediaType},
		{"missing link", "GET", "/api/links/nothere", "", http.StatusNotFound},
		{"missing redirect", "GET", "/nothere", "", http.StatusNotFound},
		{"wrong method", "PUT", "/api/links/taken", "", http.StatusMethodNotAllowed},
		{"delete missing", "DELETE", "/api/links/nothere", "", http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := e.do(tt.method, tt.path, tt.body)
			if rec.Code != tt.want {
				t.Errorf("status %d, want %d: %s", rec.Code, tt.want, rec.Body)
			}
		})
	}
}

// failingStore simulates a broken database.
type failingStore struct{}

func (failingStore) Create(context.Context, link.Link) error {
	return errors.New("connection refused: db.internal:5432")
}
func (failingStore) Get(context.Context, string) (link.Link, error) {
	return link.Link{}, errors.New("connection refused: db.internal:5432")
}
func (failingStore) Delete(context.Context, string) error {
	return errors.New("connection refused: db.internal:5432")
}

func TestServerErrorsAreNotLeaked(t *testing.T) {
	e := newEnv(t, failingStore{}, nil)
	rec := e.do("POST", "/api/links", `{"url":"https://go.dev","code":"abc"}`)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "db.internal") {
		t.Errorf("internal error leaked to the client: %s", rec.Body)
	}
	var p problem
	json.Unmarshal(rec.Body.Bytes(), &p)
	if p.RequestID == "" || p.RequestID != rec.Header().Get("X-Request-ID") {
		t.Errorf("problem should carry the request ID: %+v", p)
	}
	if !strings.Contains(e.logs.String(), "db.internal") {
		t.Errorf("the real error should be logged:\n%s", e.logs)
	}
}

func TestRateLimit(t *testing.T) {
	e := newEnv(t, link.NewMemoryStore(), NewClientLimiter(1, 3))
	codes := []int{}
	for range 5 {
		codes = append(codes, e.do("GET", "/nothere", "").Code)
	}
	if fmt.Sprint(codes) != "[404 404 404 429 429]" {
		t.Errorf("codes = %v, want three 404s then 429s", codes)
	}
	// Health checks are never rate limited.
	for range 10 {
		if c := e.do("GET", "/healthz", "").Code; c != http.StatusOK {
			t.Fatalf("healthz: %d", c)
		}
	}
}

func TestReadiness(t *testing.T) {
	var ready atomic.Bool
	h := New(Options{Service: link.NewService(link.NewMemoryStore()), Logger: slog.New(slog.DiscardHandler), Ready: &ready})
	check := func() int {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", "/readyz", nil))
		return rec.Code
	}
	if check() != http.StatusServiceUnavailable {
		t.Error("not ready yet: want 503")
	}
	ready.Store(true)
	if check() != http.StatusOK {
		t.Error("ready: want 200")
	}
}

func TestMetricsAndLogs(t *testing.T) {
	e := newEnv(t, link.NewMemoryStore(), nil)
	e.do("POST", "/api/links", `{"url":"https://go.dev","code":"gogo"}`)
	e.do("GET", "/gogo", "")
	e.do("GET", "/gogo", "")
	e.do("GET", "/missing", "")

	rec := httptest.NewRecorder()
	promhttp.HandlerFor(e.reg, promhttp.HandlerOpts{}).ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))
	for _, want := range []string{
		`shortener_http_requests_total{code="201",route="POST /api/links"} 1`,
		`shortener_http_requests_total{code="302",route="GET /{code}"} 2`,
		`shortener_http_requests_total{code="404",route="GET /{code}"} 1`,
	} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("metrics missing %q", want)
		}
	}
	if !strings.Contains(e.logs.String(), `"route":"GET /{code}","status":302`) {
		t.Errorf("access log missing route and status:\n%s", e.logs)
	}
}

func TestDelete(t *testing.T) {
	e := newEnv(t, link.NewMemoryStore(), nil)
	e.do("POST", "/api/links", `{"url":"https://go.dev","code":"temp"}`)
	if rec := e.do("DELETE", "/api/links/temp", ""); rec.Code != http.StatusNoContent || rec.Body.Len() != 0 {
		t.Fatalf("delete: %d %q", rec.Code, rec.Body)
	}
	if rec := e.do("GET", "/temp", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("redirect after delete: %d", rec.Code)
	}
}
