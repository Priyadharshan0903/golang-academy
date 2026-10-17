package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func testSite() fstest.MapFS {
	return fstest.MapFS{
		"index.html":             {Data: []byte("home")},
		"learn/index.html":       {Data: []byte("learn")},
		"css/site.css":           {Data: []byte("body{}")},
		"404.html":               {Data: []byte("custom not found")},
		"empty-dir/.placeholder": {Data: []byte("")},
	}
}

func get(t *testing.T, h http.Handler, target string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

func TestHandler(t *testing.T) {
	h := newHandler(testSite())

	tests := []struct {
		path      string
		wantCode  int
		wantBody  string
		wantCache string
	}{
		{"/healthz", 200, "ok", ""},
		{"/", 200, "home", "no-cache"},
		{"/learn/", 200, "learn", "no-cache"},
		{"/css/site.css", 200, "body{}", "public, max-age=86400"},
		{"/missing", 404, "custom not found", ""},
		{"/empty-dir/", 404, "custom not found", ""},
		{"/etc/passwd", 404, "custom not found", ""},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rec := get(t, h, tt.path)
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantCode)
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("body = %q, want it to contain %q", rec.Body.String(), tt.wantBody)
			}
			if got := rec.Header().Get("Cache-Control"); got != tt.wantCache {
				t.Errorf("Cache-Control = %q, want %q", got, tt.wantCache)
			}
		})
	}
}
