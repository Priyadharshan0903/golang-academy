package routing

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

// [start example]

func show(h http.Handler, method, path string) {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	body := strings.TrimSpace(rec.Body.String())
	if allow := rec.Header().Get("Allow"); allow != "" {
		body += " (Allow: " + allow + ")"
	}
	fmt.Printf("%-6s %-22s %d %s\n", method, path, rec.Code, body)
}

func Example_routing() {
	mux := NewMux()
	show(mux, "GET", "/")
	show(mux, "GET", "/nope")
	show(mux, "GET", "/items/42")
	show(mux, "GET", "/items/new")
	show(mux, "POST", "/items")
	show(mux, "DELETE", "/items")
	show(mux, "GET", "/files/docs/a%2Fb.txt")
	show(mux, "GET", "/static/css/site.css")
	// Output:
	// GET    /                      200 home
	// GET    /nope                  404 404 page not found
	// GET    /items/42              200 item 42
	// GET    /items/new             200 form for a new item
	// POST   /items                 201 created
	// DELETE /items                 405 Method Not Allowed (Allow: POST)
	// GET    /files/docs/a%2Fb.txt  200 file "docs/a/b.txt"
	// GET    /static/css/site.css   200 static /static/css/site.css
}

// [end example]

// [start conflict]

func Example_conflict() {
	defer func() { fmt.Println("panic:", recover() != nil) }()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /posts/{id}", func(http.ResponseWriter, *http.Request) {})
	mux.HandleFunc("GET /{resource}/latest", func(http.ResponseWriter, *http.Request) {})
	// Output: panic: true
}

// [end conflict]
