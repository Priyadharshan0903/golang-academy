// Package httpapi is the shortener's HTTP layer: routes, JSON
// encoding, error mapping and middleware. It turns HTTP requests into
// calls on link.Service and knows nothing about storage.
package httpapi

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/link"
)

// Options configures New.
type Options struct {
	Service *link.Service
	BaseURL *url.URL // public URL of the service, for short_url
	Logger  *slog.Logger
	Ready   *atomic.Bool // readiness, reported by /readyz
	Metrics *Metrics     // may be nil
	Limiter *ClientLimiter
	// ClientIP extracts the client's address for rate limiting.
	ClientIP func(*http.Request) string
}

// linkResponse is the JSON form of a link.
type linkResponse struct {
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	ShortURL  string    `json:"short_url"`
	CreatedAt time.Time `json:"created_at"`
}

type createRequest struct {
	URL  string `json:"url"`
	Code string `json:"code,omitzero"`
}

type api struct{ Options }

// [start new]

// New returns the complete HTTP handler: routes wrapped in middleware.
func New(o Options) http.Handler {
	a := &api{o}
	limit := func(h http.HandlerFunc) http.Handler {
		if o.Limiter == nil {
			return h
		}
		return o.Limiter.Middleware(o.ClientIP, h)
	}

	mux := http.NewServeMux()
	mux.Handle("POST /api/links", limit(a.createLink))
	mux.Handle("GET /api/links/{code}", limit(a.getLink))
	mux.Handle("DELETE /api/links/{code}", limit(a.deleteLink))
	mux.Handle("GET /{code}", limit(a.redirect))
	mux.HandleFunc("GET /{$}", a.home)
	mux.HandleFunc("GET /healthz", a.healthz)
	mux.HandleFunc("GET /readyz", a.readyz)

	// Outermost first. Recover sits inside Logging and Metrics so that
	// panicking requests are still logged and counted, as 500s.
	var h http.Handler = mux
	h = recoverPanics(o.Logger, h)
	if o.Metrics != nil {
		h = o.Metrics.Middleware(h)
	}
	h = logRequests(o.Logger, h)
	h = requestID(h)
	return h
}

// [end new]

// [start handlers]

func (a *api) createLink(w http.ResponseWriter, r *http.Request) {
	in, err := decodeJSON[createRequest](w, r, 16<<10)
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	l, err := a.Service.Shorten(r.Context(), in.URL, in.Code)
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/links/"+l.Code)
	writeJSON(w, http.StatusCreated, a.response(l))
}

func (a *api) getLink(w http.ResponseWriter, r *http.Request) {
	l, err := a.Service.Resolve(r.Context(), r.PathValue("code"))
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, a.response(l))
}

func (a *api) deleteLink(w http.ResponseWriter, r *http.Request) {
	if err := a.Service.Delete(r.Context(), r.PathValue("code")); err != nil {
		a.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) redirect(w http.ResponseWriter, r *http.Request) {
	l, err := a.Service.Resolve(r.Context(), r.PathValue("code"))
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	// 302, not 301: browsers cache 301s forever, so a link could never
	// be changed or counted again.
	http.Redirect(w, r, l.URL, http.StatusFound)
}

// [end handlers]

func (a *api) home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "shortener: POST /api/links with {\"url\": \"https://...\"}")
}

// healthz is the liveness check: the process is up and serving.
func (a *api) healthz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "ok")
}

// readyz is the readiness check: send traffic here. It turns false
// when shutdown begins, so load balancers stop sending new requests.
func (a *api) readyz(w http.ResponseWriter, r *http.Request) {
	if a.Ready == nil || !a.Ready.Load() {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}
	fmt.Fprintln(w, "ready")
}

func (a *api) response(l link.Link) linkResponse {
	return linkResponse{
		Code:      l.Code,
		URL:       l.URL,
		ShortURL:  a.BaseURL.JoinPath(l.Code).String(),
		CreatedAt: l.CreatedAt,
	}
}
