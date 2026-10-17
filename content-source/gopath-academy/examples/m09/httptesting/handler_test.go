package httptesting

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// [start recorder]

// Unit test: call the handler directly with a ResponseRecorder.
// No network, no server, very fast.
func TestGreetHandler(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		target     string
		wantStatus int
		wantBody   string
	}{
		{"ok", http.MethodGet, "/greet?name=Ana", 200, `{"greeting":"Hello, Ana!"}` + "\n"},
		{"missing name", http.MethodGet, "/greet", 400, `{"error":"name is required"}` + "\n"},
		{"wrong method", http.MethodPost, "/greet?name=Ana", 405, "method not allowed\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.target, nil)
			rec := httptest.NewRecorder()

			GreetHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if rec.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

// [end recorder]

// [start server]

// Integration-style test: a real HTTP server on localhost, exercised
// with a real client, including routing and headers.
func TestGreetOverHTTP(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/greet", GreetHandler)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL + "/greet?name=Bo")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["greeting"] != "Hello, Bo!" {
		t.Errorf("greeting = %q", body["greeting"])
	}
	io.Copy(io.Discard, resp.Body)
}

// [end server]
