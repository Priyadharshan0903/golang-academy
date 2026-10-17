package security

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// [start slowloris]

// A "slowloris" client opens a connection and sends its headers very
// slowly, tying up the server. ReadHeaderTimeout cuts it off.
func TestReadHeaderTimeout(t *testing.T) {
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Config.ReadHeaderTimeout = 200 * time.Millisecond
	srv.Start()
	defer srv.Close()

	start := time.Now() // before dialing: the server's clock starts when it accepts
	conn, err := net.Dial("tcp", srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	io.WriteString(conn, "GET / HTTP/1.1\r\nHost: example.com\r\n") // ...and never finish
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err = bufio.NewReader(conn).ReadString('\n')
	elapsed := time.Since(start)

	t.Logf("server gave up after %v (err=%v)", elapsed.Round(10*time.Millisecond), err)
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		t.Fatal("the server never closed the connection")
	}
	if elapsed < 200*time.Millisecond {
		t.Fatalf("closed after %v, before the timeout", elapsed)
	}
}

// [end slowloris]

func TestHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	Headers(http.NotFoundHandler()).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" || rec.Header().Get("Strict-Transport-Security") != "" {
		t.Fatalf("headers: %v", rec.Header())
	}
	rec = httptest.NewRecorder()
	Headers(http.NotFoundHandler()).ServeHTTP(rec, httptest.NewRequest("GET", "https://example.com/", nil))
	if rec.Header().Get("Strict-Transport-Security") == "" {
		t.Fatal("HSTS missing on a TLS request")
	}
}

func TestLimitBody(t *testing.T) {
	h := LimitBody(10, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.ReadAll(r.Body); err != nil {
			var mbe *http.MaxBytesError
			if errors.As(err, &mbe) {
				http.Error(w, "too large", http.StatusRequestEntityTooLarge)
				return
			}
		}
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/", strings.NewReader(strings.Repeat("x", 11))))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestNewServerHasTimeouts(t *testing.T) {
	s := NewServer(":0", http.NotFoundHandler())
	if s.ReadHeaderTimeout == 0 || s.ReadTimeout == 0 || s.WriteTimeout == 0 || s.IdleTimeout == 0 {
		t.Fatal("all timeouts must be set")
	}
}
