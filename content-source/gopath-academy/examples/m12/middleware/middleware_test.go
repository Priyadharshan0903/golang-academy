package middleware

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// [start order]

func trace(name string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("enter", name)
			next.ServeHTTP(w, r)
			fmt.Println("leave", name)
		})
	}
}

func ExampleChain() {
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("handler")
	}), trace("A"), trace("B"))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	// Output:
	// enter A
	// enter B
	// handler
	// leave B
	// leave A
}

// [end order]

// [start logtest]

// testLogger writes JSON logs without the time and duration, so the
// output is the same on every run.
func testLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey || a.Key == "duration" || a.Key == "stack" {
				return slog.Attr{}
			}
			return a
		},
	}))
}

func ExampleLogging() {
	var logs bytes.Buffer
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello/{name}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello, %s\n", r.PathValue("name"))
	})
	mux.HandleFunc("GET /boom", func(w http.ResponseWriter, r *http.Request) {
		panic("something broke")
	})
	h := Chain(mux, RequestID, Logging(testLogger(&logs)), Recover(testLogger(&logs)))

	for _, path := range []string{"/hello/gopher", "/boom"} {
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("X-Request-ID", "req-1")
		h.ServeHTTP(httptest.NewRecorder(), req)
	}
	fmt.Print(logs.String())
	// Output:
	// {"level":"INFO","msg":"request","method":"GET","path":"/hello/gopher","pattern":"GET /hello/{name}","status":200,"bytes":14,"request_id":"req-1"}
	// {"level":"ERROR","msg":"panic in handler","panic":"something broke","request_id":"req-1"}
	// {"level":"INFO","msg":"request","method":"GET","path":"/boom","pattern":"GET /boom","status":500,"bytes":22,"request_id":"req-1"}
}

// [end logtest]

func TestRequestIDGeneratedWhenInvalid(t *testing.T) {
	var seen string
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = RequestIDFrom(r.Context())
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "bad\nid")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if len(seen) != 16 || strings.Contains(seen, "\n") || rec.Header().Get("X-Request-ID") != seen {
		t.Fatalf("got id %q, header %q", seen, rec.Header().Get("X-Request-ID"))
	}
}

func TestResponseControllerThroughRecorder(t *testing.T) {
	h := Logging(slog.New(slog.DiscardHandler))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("partial"))
		if err := http.NewResponseController(w).Flush(); err != nil {
			t.Errorf("Flush through the wrapper: %v", err)
		}
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if !rec.Flushed {
		t.Fatal("expected the response to be flushed")
	}
}
