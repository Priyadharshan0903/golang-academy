// Package security holds the examples for lesson 12.6.
package security

import (
	"crypto/tls"
	"net/http"
	"time"
)

// [start server]

// NewServer returns an http.Server with limits suitable for a service
// exposed to the internet. The zero values in http.Server mean "no
// limit", so every one of these must be set explicitly.
func NewServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: h,

		// Slow clients must not hold connections open forever.
		ReadHeaderTimeout: 5 * time.Second,   // time to read the request headers
		ReadTimeout:       30 * time.Second,  // time to read the whole request, body included
		WriteTimeout:      60 * time.Second,  // time to write the response
		IdleTimeout:       120 * time.Second, // keep-alive connections waiting for the next request

		MaxHeaderBytes: 64 << 10, // 64 KiB of headers is plenty (the default is 1 MiB)

		TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12},
	}
}

// [end server]

// [start headers]

// Headers sets response headers that make browsers safer when they
// talk to this service.
func Headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff") // never guess a different content type
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'") // an API serves no pages
		if r.TLS != nil {
			h.Set("Strict-Transport-Security", "max-age=31536000") // only ever use HTTPS from now on
		}
		next.ServeHTTP(w, r)
	})
}

// [end headers]

// [start limit]

// LimitBody caps every request body at n bytes. Reading more fails
// with an *http.MaxBytesError, and the server closes the connection.
func LimitBody(n int64, next http.Handler) http.Handler {
	return http.MaxBytesHandler(next, n)
}

// [end limit]
