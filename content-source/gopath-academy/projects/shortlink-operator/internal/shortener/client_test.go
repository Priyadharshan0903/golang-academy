package shortener

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			http.Error(w, `{"detail":"link not found"}`, http.StatusNotFound)
		case r.Method == http.MethodPost:
			http.Error(w, `{"detail":"code already in use"}`, http.StatusConflict)
		default:
			http.Error(w, `{"detail":"boom"}`, http.StatusInternalServerError)
		}
	}))
	defer srv.Close()
	c := New(srv.URL)
	ctx := context.Background()
	if _, err := c.Get(ctx, "x"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get: %v", err)
	}
	if _, err := c.Create(ctx, "x", "https://go.dev"); !errors.Is(err, ErrConflict) {
		t.Errorf("Create: %v", err)
	}
	if err := c.Delete(ctx, "x"); err == nil || errors.Is(err, ErrNotFound) {
		t.Errorf("Delete: %v", err)
	}
}
