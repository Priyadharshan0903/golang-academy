package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"time"
)

// [start client]

// One client for the whole program: it pools and reuses connections.
var client = &http.Client{
	Timeout: 10 * time.Second, // whole request: connect, headers and body
}

type Repo struct {
	Name  string `json:"name"`
	Stars int    `json:"stars"`
}

// getRepo shows every step a correct request needs.
func getRepo(ctx context.Context, baseURL, name string) (Repo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/repos/"+name, nil)
	if err != nil {
		return Repo{}, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return Repo{}, fmt.Errorf("get repo %s: %w", name, err)
	}
	defer resp.Body.Close() // always close, or the connection cannot be reused

	if resp.StatusCode != http.StatusOK { // a 404 or 500 is NOT an error from Do
		io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return Repo{}, fmt.Errorf("get repo %s: unexpected status %s", name, resp.Status)
	}

	var r Repo
	// Limit how much we are willing to read from someone else's server.
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&r); err != nil {
		return Repo{}, fmt.Errorf("decode repo %s: %w", name, err)
	}
	return r, nil
}

// [end client]

// [start test]

func Example_request() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/go" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"name":"go","stars":125000}`)
	}))
	defer srv.Close()

	ctx := context.Background()
	repo, err := getRepo(ctx, srv.URL, "go")
	fmt.Println(repo, err)

	_, err = getRepo(ctx, srv.URL, "nope")
	fmt.Println(err)

	// Output:
	// {go 125000} <nil>
	// get repo nope: unexpected status 404 Not Found
}

// [end test]

// [start timeout]

func Example_timeouts() {
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(time.Second):
		case <-r.Context().Done(): // the client gave up
		}
	}))
	defer slow.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := getRepo(ctx, slow.URL, "go")

	var netErr net.Error
	fmt.Println(errors.Is(err, context.DeadlineExceeded))
	fmt.Println(errors.As(err, &netErr) && netErr.Timeout())

	// Output:
	// true
	// true
}

// [end timeout]

// [start transport]

// newTunedClient shows the Transport settings worth knowing. The
// defaults are reasonable; tune them when profiling or metrics say so.
func newTunedClient() *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxIdleConnsPerHost = 20 // default is 2: low for services calling one backend a lot
	t.IdleConnTimeout = 90 * time.Second
	t.TLSHandshakeTimeout = 5 * time.Second
	t.ResponseHeaderTimeout = 5 * time.Second
	return &http.Client{Transport: t, Timeout: 15 * time.Second}
}

func Example_transport() {
	c := newTunedClient()
	fmt.Println(c.Timeout, c.Transport.(*http.Transport).MaxIdleConnsPerHost)

	// Output:
	// 15s 20
}

// [end transport]
