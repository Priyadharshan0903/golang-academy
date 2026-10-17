package observability

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		// A child span for work inside the handler.
		_, span := trace.SpanFromContext(r.Context()).TracerProvider().Tracer("items").Start(r.Context(), "load item")
		defer span.End()
		fmt.Fprintf(w, "item %s", r.PathValue("id"))
	})
	return mux
}

// [start metricstest]

func TestMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	h := NewMetrics(reg).Middleware(newMux())
	for _, path := range []string{"/items/1", "/items/2", "/items/3", "/nope"} {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", path, nil))
	}

	// Scrape the registry the way Prometheus would.
	rec := httptest.NewRecorder()
	promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))
	out := rec.Body.String()
	for _, want := range []string{
		`http_requests_total{code="200",route="GET /items/{id}"} 3`,
		`http_requests_total{code="404",route="unmatched"} 1`,
		`http_request_duration_seconds_count{route="GET /items/{id}"} 3`,
		`http_requests_in_flight 0`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "http_requests_total") || strings.HasPrefix(line, "# TYPE http_requests_total") {
			t.Log(line)
		}
	}
}

// [end metricstest]

// [start tracetest]

func TestTracing(t *testing.T) {
	spans := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans))
	defer tp.Shutdown(context.Background())

	h := Tracing(tp, newMux())
	req := httptest.NewRequest("GET", "/items/42", nil)
	// The caller's trace context, as another service would send it.
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	h.ServeHTTP(httptest.NewRecorder(), req)

	ended := spans.Ended()
	if len(ended) != 2 {
		t.Fatalf("got %d spans, want 2", len(ended))
	}
	child, server := ended[0], ended[1] // the child ends first
	t.Logf("server span %q trace=%s parent=%s attrs=%v", server.Name(), server.SpanContext().TraceID(), server.Parent().SpanID(), server.Attributes())
	t.Logf("child span %q parent=%s", child.Name(), child.Parent().SpanID())

	if server.Name() != "GET /items/{id}" {
		t.Errorf("server span name = %q", server.Name())
	}
	if got := server.SpanContext().TraceID().String(); got != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Errorf("trace ID = %s, want the caller's", got)
	}
	if server.Parent().SpanID().String() != "00f067aa0ba902b7" {
		t.Errorf("server span's parent = %s, want the caller's span", server.Parent().SpanID())
	}
	if child.Parent().SpanID() != server.SpanContext().SpanID() {
		t.Error("child span is not a child of the server span")
	}
}

// [end tracetest]
