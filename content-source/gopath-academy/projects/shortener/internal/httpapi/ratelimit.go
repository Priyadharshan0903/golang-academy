package httpapi

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// ClientLimiter keeps one token bucket per client (lesson 7.5 built one
// by hand; this uses golang.org/x/time/rate).
type ClientLimiter struct {
	mu      sync.Mutex
	clients map[string]*client
	rate    rate.Limit
	burst   int
}

type client struct {
	lim      *rate.Limiter
	lastSeen time.Time
}

// NewClientLimiter allows each client perSecond requests per second on
// average, with bursts of up to burst requests.
func NewClientLimiter(perSecond float64, burst int) *ClientLimiter {
	return &ClientLimiter{clients: map[string]*client{}, rate: rate.Limit(perSecond), burst: burst}
}

// Allow reports whether the client identified by key may proceed now.
func (l *ClientLimiter) Allow(key string) bool {
	l.mu.Lock()
	c, ok := l.clients[key]
	if !ok {
		c = &client{lim: rate.NewLimiter(l.rate, l.burst)}
		l.clients[key] = c
	}
	c.lastSeen = time.Now()
	l.mu.Unlock()
	return c.lim.Allow()
}

// Prune forgets clients not seen for idle, so the map does not grow
// forever. It returns how many were removed. Note that this keeps the
// map's memory (lesson 11.6), which is fine for a bounded client set.
func (l *ClientLimiter) Prune(idle time.Duration) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	n := 0
	for k, c := range l.clients {
		if time.Since(c.lastSeen) > idle {
			delete(l.clients, k)
			n++
		}
	}
	return n
}

// Middleware rejects requests over the limit with 429.
func (l *ClientLimiter) Middleware(clientIP func(*http.Request) string, next http.Handler) http.Handler {
	if clientIP == nil {
		clientIP = RemoteIP
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow(clientIP(r)) {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RemoteIP is the address of the TCP peer: the client, or the last
// proxy in front of the service.
func RemoteIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// HeaderIP returns a ClientIP function that trusts the given header,
// for example X-Real-IP set by an ingress controller. Only use it when
// every request passes through a proxy that sets the header, or
// clients can pick their own rate limit key.
func HeaderIP(header string) func(*http.Request) string {
	return func(r *http.Request) string {
		if v := r.Header.Get(header); v != "" {
			return v
		}
		return RemoteIP(r)
	}
}
