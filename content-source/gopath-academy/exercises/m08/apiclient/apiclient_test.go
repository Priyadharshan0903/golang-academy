package apiclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/json" {
			http.Error(w, `{"error":"missing Accept header"}`, http.StatusNotAcceptable)
			return
		}
		switch r.PathValue("id") {
		case "7":
			w.Write([]byte(`{"id":7,"name":"Ana","email":"ana@example.com"}`))
		case "500":
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"maintenance"}`))
		case "999":
			select { // a slow endpoint
			case <-time.After(2 * time.Second):
			case <-r.Context().Done():
			}
		default:
			http.NotFound(w, r)
		}
	})
	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
		var in struct{ Name, Email string }
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"name is required"}`))
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": 8, "name": in.Name, "email": in.Email})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func newClient(srv *httptest.Server) *Client {
	return &Client{BaseURL: srv.URL, HTTP: srv.Client()}
}

func TestGetUser(t *testing.T) {
	c := newClient(newServer(t))
	u, err := c.GetUser(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if u != (User{ID: 7, Name: "Ana", Email: "ana@example.com"}) {
		t.Errorf("got %+v", u)
	}
}

func TestGetUserNotFound(t *testing.T) {
	c := newClient(newServer(t))
	_, err := c.GetUser(context.Background(), 404)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestAPIError(t *testing.T) {
	c := newClient(newServer(t))
	_, err := c.GetUser(context.Background(), 500)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.Status != 503 || apiErr.Message != "maintenance" {
		t.Errorf("got %+v", apiErr)
	}
	if apiErr.Error() == "" {
		t.Error("APIError.Error() is empty")
	}
}

func TestContextRespected(t *testing.T) {
	c := newClient(newServer(t))

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, err := c.GetUser(ctx, 999); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("slow request: err = %v, want context.DeadlineExceeded", err)
	}
	if time.Since(start) > time.Second {
		t.Error("the request was not cancelled at the deadline")
	}

	cancelled, cancel2 := context.WithCancel(context.Background())
	cancel2()
	if _, err := c.GetUser(cancelled, 7); !errors.Is(err, context.Canceled) {
		t.Errorf("cancelled context: err = %v, want context.Canceled", err)
	}
}

func TestCreateUser(t *testing.T) {
	c := newClient(newServer(t))
	u, err := c.CreateUser(context.Background(), "Bo", "bo@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != 8 || u.Name != "Bo" {
		t.Errorf("got %+v", u)
	}
	_, err = c.CreateUser(context.Background(), "", "x@example.com")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != 400 || apiErr.Message != "name is required" {
		t.Errorf("err = %v, want APIError 400 'name is required'", err)
	}
}
