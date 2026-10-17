package fizzhttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

func call(method, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	Handler(rec, httptest.NewRequest(method, target, nil))
	return rec
}

func TestOK(t *testing.T) {
	rec := call(http.MethodGet, "/fizzbuzz?n=15")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
	var body struct {
		N      int      `json:"n"`
		Result []string `json:"result"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v\n%s", err, rec.Body.String())
	}
	want := []string{"1", "2", "Fizz", "4", "Buzz", "Fizz", "7", "8", "Fizz", "Buzz", "11", "Fizz", "13", "14", "FizzBuzz"}
	if body.N != 15 || !slices.Equal(body.Result, want) {
		t.Errorf("got %+v", body)
	}
}

func TestBadInput(t *testing.T) {
	for _, target := range []string{"/fizzbuzz", "/fizzbuzz?n=abc", "/fizzbuzz?n=0", "/fizzbuzz?n=1001", "/fizzbuzz?n=-3"} {
		rec := call(http.MethodGet, target)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", target, rec.Code)
			continue
		}
		var e map[string]string
		if json.Unmarshal(rec.Body.Bytes(), &e) != nil || e["error"] == "" {
			t.Errorf("%s: body %q should be JSON with an error field", target, rec.Body.String())
		}
	}
}

func TestMethodNotAllowed(t *testing.T) {
	rec := call(http.MethodPost, "/fizzbuzz?n=3")
	if rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != "GET" {
		t.Errorf("status = %d, Allow = %q", rec.Code, rec.Header().Get("Allow"))
	}
}

func TestThroughRealServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(Handler))
	defer srv.Close()
	resp, err := srv.Client().Get(srv.URL + "/fizzbuzz?n=3")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("status = %d", resp.StatusCode)
	}
}
