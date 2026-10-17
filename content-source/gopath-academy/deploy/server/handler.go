package main

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// newHandler returns the HTTP handler for the site stored in site.
func newHandler(site fs.FS) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/", staticHandler(site))
	return mux
}

// staticHandler serves files from site and falls back to 404.html.
func staticHandler(site fs.FS) http.Handler {
	files := http.FileServerFS(site)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !exists(site, r.URL.Path) {
			serveNotFound(w, site)
			return
		}
		setCacheHeaders(w, r.URL.Path)
		files.ServeHTTP(w, r)
	})
}

// exists reports whether urlPath maps to a file, or to a directory
// containing index.html.
func exists(site fs.FS, urlPath string) bool {
	name := strings.TrimPrefix(path.Clean("/"+urlPath), "/")
	if name == "" {
		name = "."
	}
	fi, err := fs.Stat(site, name)
	if err != nil {
		return false
	}
	if fi.IsDir() {
		_, err = fs.Stat(site, path.Join(name, "index.html"))
		return err == nil
	}
	return true
}

func serveNotFound(w http.ResponseWriter, site fs.FS) {
	body, err := fs.ReadFile(site, "404.html")
	if err != nil {
		http.NotFound(w, nil)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write(body)
}

// setCacheHeaders lets browsers cache fingerprinted assets for a long
// time while always revalidating HTML pages.
func setCacheHeaders(w http.ResponseWriter, urlPath string) {
	switch path.Ext(urlPath) {
	case ".css", ".js", ".woff2", ".png", ".svg", ".jpg", ".webp", ".ico":
		w.Header().Set("Cache-Control", "public, max-age=86400")
	default:
		w.Header().Set("Cache-Control", "no-cache")
	}
}
