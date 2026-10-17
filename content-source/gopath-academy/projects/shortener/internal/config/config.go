// Package config loads the shortener's configuration from the
// environment, following the twelve-factor approach of lesson 12.4.
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

// Config holds every setting. See the README for the variables.
type Config struct {
	Addr            string        // ADDR: public API listener
	AdminAddr       string        // ADMIN_ADDR: metrics and pprof listener, keep it internal
	BaseURL         *url.URL      // BASE_URL: used to build short URLs
	LogLevel        slog.Level    // LOG_LEVEL: debug, info, warn, error
	LogFormat       string        // LOG_FORMAT: json or text
	RateLimit       float64       // RATE_LIMIT: requests per second per client
	RateBurst       int           // RATE_BURST: short bursts allowed above the rate
	ClientIPHeader  string        // CLIENT_IP_HEADER: e.g. X-Real-IP, only behind a proxy that sets it
	ShutdownDelay   time.Duration // SHUTDOWN_DELAY: time to stay up, unready, before shutting down
	ShutdownTimeout time.Duration // SHUTDOWN_TIMEOUT: time allowed for in-flight requests to finish
	DatabaseURL     string        // DATABASE_URL or DATABASE_URL_FILE: PostgreSQL; empty means in memory
	Migrate         bool          // MIGRATE: apply database migrations at start-up
	RedisAddr       string        // REDIS_ADDR: enables the cache when set
	CacheTTL        time.Duration // CACHE_TTL: how long links stay cached
}

// Load reads the configuration through getenv and reports every
// problem at once.
func Load(getenv func(string) string) (Config, error) {
	c := Config{
		Addr:            ":8080",
		AdminAddr:       ":9090",
		LogLevel:        slog.LevelInfo,
		LogFormat:       "json",
		RateLimit:       5,
		RateBurst:       10,
		ShutdownTimeout: 10 * time.Second,
		Migrate:         true,
		CacheTTL:        time.Hour,
	}
	var errs []error
	str := func(name string, dst *string) {
		if v := getenv(name); v != "" {
			*dst = v
		}
	}
	dur := func(name string, dst *time.Duration) {
		if v := getenv(name); v != "" {
			d, err := time.ParseDuration(v)
			if err != nil || d < 0 {
				errs = append(errs, fmt.Errorf("%s must be a duration like 5s, got %q", name, v))
			}
			*dst = d
		}
	}

	str("ADDR", &c.Addr)
	str("ADMIN_ADDR", &c.AdminAddr)
	str("LOG_FORMAT", &c.LogFormat)
	str("CLIENT_IP_HEADER", &c.ClientIPHeader)
	dur("SHUTDOWN_DELAY", &c.ShutdownDelay)
	dur("SHUTDOWN_TIMEOUT", &c.ShutdownTimeout)
	dur("CACHE_TTL", &c.CacheTTL)
	str("REDIS_ADDR", &c.RedisAddr)
	if v := getenv("MIGRATE"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			errs = append(errs, fmt.Errorf("MIGRATE must be true or false, got %q", v))
		}
		c.Migrate = b
	}
	dbURL, err := secret(getenv, "DATABASE_URL")
	if err != nil {
		errs = append(errs, err)
	}
	c.DatabaseURL = dbURL

	base := getenv("BASE_URL")
	if base == "" {
		base = "http://localhost:8080"
	}
	if u, err := url.Parse(base); err != nil || u.Scheme == "" || u.Host == "" {
		errs = append(errs, fmt.Errorf("BASE_URL must be an absolute URL, got %q", base))
	} else {
		c.BaseURL = u
	}
	if v := getenv("LOG_LEVEL"); v != "" {
		if err := c.LogLevel.UnmarshalText([]byte(v)); err != nil {
			errs = append(errs, fmt.Errorf("LOG_LEVEL: %w", err))
		}
	}
	if c.LogFormat != "json" && c.LogFormat != "text" {
		errs = append(errs, fmt.Errorf("LOG_FORMAT must be json or text, got %q", c.LogFormat))
	}
	if v := getenv("RATE_LIMIT"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || f <= 0 {
			errs = append(errs, fmt.Errorf("RATE_LIMIT must be a positive number, got %q", v))
		}
		c.RateLimit = f
	}
	if v := getenv("RATE_BURST"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			errs = append(errs, fmt.Errorf("RATE_BURST must be a positive integer, got %q", v))
		}
		c.RateBurst = n
	}
	return c, errors.Join(errs...)
}

// secret reads NAME, or the file named by NAME_FILE (lesson 12.4).
func secret(getenv func(string) string, name string) (string, error) {
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
