package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunCountsResults(t *testing.T) {
	var n atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != "app.example" {
			t.Errorf("Host = %q", r.Host)
		}
		if n.Add(1)%4 == 0 {
			w.WriteHeader(http.StatusBadGateway)
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	r := run(ctx, srv.URL, "app.example", 2)
	if r.total == 0 || r.statuses[502] == 0 || r.failed() != r.statuses[502] {
		t.Fatalf("unexpected result:\n%s", r)
	}
	t.Log(r)
}
