package observability

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// [start tracing]

// Tracing starts a server span for every request. If the caller sent a
// W3C traceparent header, the span joins the caller's trace.
func Tracing(tp trace.TracerProvider, mux http.Handler) http.Handler {
	tracer := tp.Tracer("github.com/Priyadharshan0903/gopath-academy/examples/thirdparty/observability")
	prop := propagation.TraceContext{}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := prop.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := tracer.Start(ctx, r.Method, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		r = r.WithContext(ctx) // handlers can now start child spans from r.Context()
		mux.ServeHTTP(rec, r)

		// The route is known only after the mux has matched the request,
		// and the mux records it on the request it was given.
		if r.Pattern != "" {
			span.SetName(r.Pattern)
			span.SetAttributes(attribute.String("http.route", r.Pattern))
		}
		span.SetAttributes(
			attribute.String("http.request.method", r.Method),
			attribute.Int("http.response.status_code", rec.status),
		)
		if rec.status >= 500 {
			span.SetStatus(codes.Error, http.StatusText(rec.status))
		}
	})
}

// [end tracing]
