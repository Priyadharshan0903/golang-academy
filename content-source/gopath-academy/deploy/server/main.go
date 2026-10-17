// Command server serves the built website.
//
// It is intentionally small: a static file server with a health check,
// sensible timeouts, a custom 404 page and graceful shutdown. It is the
// workload we run on Kubernetes in the K8s track.
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dir := flag.String("dir", "public", "directory with the built site")
	shutdownDelay := flag.Duration("shutdown-delay", 0, "keep serving this long after SIGTERM, while Kubernetes removes the pod from load balancing (lesson 15.2)")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if fi, err := os.Stat(*dir); err != nil || !fi.IsDir() {
		logger.Error("site directory not found", "dir", *dir)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:              *addr,
		Handler:           newHandler(os.DirFS(*dir)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Kubernetes sends SIGTERM before killing a pod. Stop accepting new
	// connections and let in-flight requests finish.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("listening", "addr", *addr, "dir", *dir)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down", "delay", *shutdownDelay)
	time.Sleep(*shutdownDelay)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}
}
