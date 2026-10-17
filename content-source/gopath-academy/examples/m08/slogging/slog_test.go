package slogging

import (
	"context"
	"log/slog"
	"os"
)

// newExampleLogger removes the time attribute so example output is stable.
func newExampleLogger(json bool, level slog.Leveler) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && len(groups) == 0 {
				return slog.Attr{} // drop it
			}
			return a
		},
	}
	if json {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}

// [start basics]

func Example_basics() {
	log := newExampleLogger(false, slog.LevelInfo)

	log.Info("server started", "port", 8080, "env", "prod")
	log.Debug("not shown: below the Info level")
	log.Warn("slow query", slog.Duration("took", 1500_000_000), slog.String("table", "orders"))

	// Output:
	// level=INFO msg="server started" port=8080 env=prod
	// level=WARN msg="slow query" took=1.5s table=orders
}

// [end basics]

// [start json]

func Example_jsonAndWith() {
	log := newExampleLogger(true, slog.LevelInfo)

	// With returns a logger that adds these attributes to every record.
	reqLog := log.With("request_id", "r-42", slog.Group("user", "id", 7))
	reqLog.Info("order created", "order_id", 1001)

	// Output:
	// {"level":"INFO","msg":"order created","request_id":"r-42","user":{"id":7},"order_id":1001}
}

// [end json]

// [start levels]

func Example_dynamicLevel() {
	var level slog.LevelVar // zero value is Info; can be changed at run time
	log := newExampleLogger(false, &level)

	log.Debug("hidden")
	level.Set(slog.LevelDebug) // e.g. toggled by an admin endpoint
	log.Debug("now visible")

	// Output:
	// level=DEBUG msg="now visible"
}

// [end levels]

// [start redact]

type Token string

// LogValue controls how a type appears in logs, wherever it is logged.
func (Token) LogValue() slog.Value { return slog.StringValue("REDACTED") }

func Example_redaction() {
	log := newExampleLogger(false, slog.LevelInfo)
	log.Info("calling api", "token", Token("sk-live-123"))

	// Output:
	// level=INFO msg="calling api" token=REDACTED
}

// [end redact]

// [start context]

type ctxKey struct{}

// ContextHandler adds a request ID from the context to every record.
// It wraps any other Handler: handlers compose like io.Writers.
type ContextHandler struct{ slog.Handler }

func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if id, ok := ctx.Value(ctxKey{}).(string); ok {
		r.AddAttrs(slog.String("request_id", id))
	}
	return h.Handler.Handle(ctx, r)
}

func Example_contextHandler() {
	base := newExampleLogger(false, slog.LevelInfo).Handler()
	log := slog.New(ContextHandler{base})

	ctx := context.WithValue(context.Background(), ctxKey{}, "r-99")
	log.InfoContext(ctx, "payment accepted", "amount", 499)

	// Output:
	// level=INFO msg="payment accepted" amount=499 request_id=r-99
}

// [end context]
