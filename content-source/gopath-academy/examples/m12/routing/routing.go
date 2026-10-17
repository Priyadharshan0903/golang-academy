// Package routing holds the examples for lesson 12.1.
package routing

import (
	"fmt"
	"net/http"
)

// [start mux]

// NewMux shows the routing patterns of http.ServeMux (Go 1.22+).
func NewMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Exact match for "/" only. Without {$}, "/" would match every path.
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "home")
	})

	// A method and a wildcard. GET also matches HEAD requests.
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "item %s\n", r.PathValue("id"))
	})

	// A more specific pattern wins, whatever the registration order.
	mux.HandleFunc("GET /items/new", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "form for a new item")
	})

	mux.HandleFunc("POST /items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, "created")
	})

	// A trailing "..." wildcard matches the rest of the path.
	mux.HandleFunc("GET /files/{path...}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "file %q\n", r.PathValue("path"))
	})

	// A pattern ending in "/" matches the whole subtree.
	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "static %s\n", r.URL.Path)
	})

	return mux
}

// [end mux]
