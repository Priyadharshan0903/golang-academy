// Package config holds the examples for lesson 12.4.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// [start config]

// Config is everything the service needs to know about its
// environment. It is loaded once at start-up and passed down
// explicitly; no package reads environment variables on its own.
type Config struct {
	Addr            string        // ADDR, default ":8080"
	BaseURL         *url.URL      // BASE_URL, required
	LogLevel        slog.Level    // LOG_LEVEL: debug, info, warn, error
	ShutdownTimeout time.Duration // SHUTDOWN_TIMEOUT, default 10s
	RateLimit       float64       // RATE_LIMIT, requests per second per client
	DatabaseURL     string        // DATABASE_URL or DATABASE_URL_FILE, a secret
}

// Load reads the configuration through getenv (os.Getenv in main, a
// map in tests) and reports every problem at once.
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		Addr:            ":8080",
		LogLevel:        slog.LevelInfo,
		ShutdownTimeout: 10 * time.Second,
		RateLimit:       10,
	}
	var errs []error

	if v := getenv("ADDR"); v != "" {
		cfg.Addr = v
	}
	if v := getenv("BASE_URL"); v == "" {
		errs = append(errs, errors.New("BASE_URL is required"))
	} else if u, err := url.Parse(v); err != nil || u.Scheme == "" || u.Host == "" {
		errs = append(errs, fmt.Errorf("BASE_URL must be an absolute URL, got %q", v))
	} else {
		cfg.BaseURL = u
	}
	if v := getenv("LOG_LEVEL"); v != "" {
		if err := cfg.LogLevel.UnmarshalText([]byte(v)); err != nil {
			errs = append(errs, fmt.Errorf("LOG_LEVEL: %w", err))
		}
	}
	if v := getenv("SHUTDOWN_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			errs = append(errs, fmt.Errorf("SHUTDOWN_TIMEOUT must be a positive duration like 15s, got %q", v))
		}
		cfg.ShutdownTimeout = d
	}
	if v := getenv("RATE_LIMIT"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || f <= 0 {
			errs = append(errs, fmt.Errorf("RATE_LIMIT must be a positive number, got %q", v))
		}
		cfg.RateLimit = f
	}

	secret, err := secretFrom(getenv, "DATABASE_URL")
	if err != nil {
		errs = append(errs, err)
	}
	cfg.DatabaseURL = secret

	return cfg, errors.Join(errs...)
}

// [end config]

// [start secret]

// secretFrom reads a secret from NAME, or from the file named by
// NAME_FILE. Files are how Kubernetes and Docker usually provide
// secrets, and they keep secrets out of `env` output and crash dumps.
func secretFrom(getenv func(string) string, name string) (string, error) {
	value, file := getenv(name), getenv(name+"_FILE")
	switch {
	case value != "" && file != "":
		return "", fmt.Errorf("set %s or %s_FILE, not both", name, name)
	case file != "":
		b, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("%s_FILE: %w", name, err)
		}
		return strings.TrimSpace(string(b)), nil
	default:
		return value, nil
	}
}

// [end secret]

// [start redact]

// LogValue makes slog print the config without the secret, so logging
// the config at start-up is safe.
func (c Config) LogValue() slog.Value {
	base := ""
	if c.BaseURL != nil {
		base = c.BaseURL.String()
	}
	return slog.GroupValue(
		slog.String("addr", c.Addr),
		slog.String("base_url", base),
		slog.String("log_level", c.LogLevel.String()),
		slog.Duration("shutdown_timeout", c.ShutdownTimeout),
		slog.Float64("rate_limit", c.RateLimit),
		slog.Bool("database_url_set", c.DatabaseURL != ""),
	)
}

// [end redact]
