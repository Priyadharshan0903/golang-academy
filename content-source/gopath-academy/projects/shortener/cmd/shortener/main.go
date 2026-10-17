// Command shortener runs the URL shortener's HTTP service.
//
// Configuration comes from environment variables; see the README.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"slices"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/cache"
	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/config"
	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/httpapi"
	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/link"
	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

func main() {
	// SIGTERM is what Kubernetes sends; SIGINT is Ctrl+C.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Getenv, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "shortener:", err)
		os.Exit(1)
	}
}

// [start wiring]

// run is the whole program. It takes its dependencies as parameters,
// so tests can run it with their own environment and output, and stop
// it by cancelling ctx.
func run(ctx context.Context, getenv func(string) string, stdout io.Writer) error {
	cfg, err := config.Load(getenv)
	if err != nil {
		return fmt.Errorf("configuration: %w", err)
	}
	logger := newLogger(cfg, stdout)
	logger.Info("starting", "addr", cfg.Addr, "admin_addr", cfg.AdminAddr, "base_url", cfg.BaseURL.String())

	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	limiter := httpapi.NewClientLimiter(cfg.RateLimit, cfg.RateBurst)
	clientIP := httpapi.RemoteIP
	if cfg.ClientIPHeader != "" {
		clientIP = httpapi.HeaderIP(cfg.ClientIPHeader)
	}

	store, closeStore, err := openStore(ctx, cfg, logger)
	if err != nil {
		return err
	}
	defer closeStore()

	var ready atomic.Bool
	api := newServer(httpapi.New(httpapi.Options{
		Service:  link.NewService(store),
		BaseURL:  cfg.BaseURL,
		Logger:   logger,
		Ready:    &ready,
		Metrics:  httpapi.NewMetrics(reg),
		Limiter:  limiter,
		ClientIP: clientIP,
	}))
	admin := newServer(adminMux(reg))

	// [end wiring]

	// Listen before serving, so a bad address fails start-up immediately
	// and port 0 (any free port, used in tests) can be logged.
	apiLn, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return err
	}
	adminLn, err := net.Listen("tcp", cfg.AdminAddr)
	if err != nil {
		apiLn.Close()
		return err
	}

	serveErr := make(chan error, 2)
	var wg sync.WaitGroup
	for _, s := range []struct {
		srv  *http.Server
		ln   net.Listener
		name string
	}{{api, apiLn, "api"}, {admin, adminLn, "admin"}} {
		wg.Go(func() {
			logger.Info("listening", "server", s.name, "addr", s.ln.Addr().String())
			if err := s.srv.Serve(s.ln); !errors.Is(err, http.ErrServerClosed) {
				serveErr <- fmt.Errorf("%s server: %w", s.name, err)
			}
		})
	}
	wg.Go(func() { pruneLoop(ctx, limiter) })
	ready.Store(true)

	// [start shutdown]

	select {
	case err = <-serveErr:
		logger.Error("server failed", "error", err)
	case <-ctx.Done():
		logger.Info("shutting down")
	}

	// Graceful shutdown (lesson 7.7). First report "not ready" and keep
	// serving for SHUTDOWN_DELAY, so a load balancer that still routes to
	// us has time to notice. Then stop accepting connections and let
	// in-flight requests finish within SHUTDOWN_TIMEOUT.
	ready.Store(false)
	if err == nil && cfg.ShutdownDelay > 0 {
		time.Sleep(cfg.ShutdownDelay)
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	shutdownErr := errors.Join(api.Shutdown(shutdownCtx), admin.Shutdown(shutdownCtx))
	wg.Wait()
	logger.Info("stopped")
	return errors.Join(err, shutdownErr)
}

// [end shutdown]

// [start store]

// openStore picks the storage from the configuration: PostgreSQL when
// DATABASE_URL is set (milestone 3), memory otherwise, with an optional
// Redis cache in front. The returned function releases connections.
func openStore(ctx context.Context, cfg config.Config, logger *slog.Logger) (link.Store, func(), error) {
	var closers []func()
	closeAll := func() {
		for _, c := range slices.Backward(closers) {
			c()
		}
	}

	var store link.Store
	if cfg.DatabaseURL == "" {
		logger.Warn("DATABASE_URL is not set: links are kept in memory and lost on restart")
		store = link.NewMemoryStore()
	} else {
		pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, nil, fmt.Errorf("database: %w", err)
		}
		closers = append(closers, pool.Close)
		pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := pool.Ping(pingCtx); err != nil {
			closeAll()
			return nil, nil, fmt.Errorf("database: %w", err)
		}
		if cfg.Migrate {
			applied, err := postgres.Migrate(ctx, pool)
			if err != nil {
				closeAll()
				return nil, nil, fmt.Errorf("migrate: %w", err)
			}
			logger.Info("migrations", "applied", applied)
		}
		store = postgres.NewStore(pool)
	}

	if cfg.RedisAddr != "" {
		rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
		closers = append(closers, func() { rdb.Close() })
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		if err := rdb.Ping(pingCtx).Err(); err != nil {
			// Start anyway: the cache only makes things faster.
			logger.Warn("redis unavailable, continuing without a warm cache", "error", err)
		}
		store = cache.New(store, rdb, cfg.CacheTTL, logger)
	}
	return store, closeAll, nil
}

// [end store]

func newLogger(cfg config.Config, w io.Writer) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}
	if cfg.LogFormat == "text" {
		return slog.New(slog.NewTextHandler(w, opts))
	}
	return slog.New(slog.NewJSONHandler(w, opts))
}

// newServer sets the timeouts from lesson 12.6.
func newServer(h http.Handler) *http.Server {
	return &http.Server{
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    64 << 10,
	}
}

// adminMux serves operational endpoints on a separate port that is
// never exposed publicly: Prometheus metrics and pprof profiles.
func adminMux(reg *prometheus.Registry) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	return mux
}

// pruneLoop forgets idle rate limit clients every minute.
func pruneLoop(ctx context.Context, l *httpapi.ClientLimiter) {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			l.Prune(10 * time.Minute)
		}
	}
}
