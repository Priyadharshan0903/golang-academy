package config

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

// [start example]

func ExampleLoad() {
	cfg, err := Load(env(map[string]string{
		"BASE_URL":         "https://sho.rt",
		"LOG_LEVEL":        "debug",
		"SHUTDOWN_TIMEOUT": "30s",
	}))
	fmt.Println(err)
	fmt.Println(cfg.Addr, cfg.BaseURL, cfg.LogLevel, cfg.ShutdownTimeout, cfg.RateLimit)

	_, err = Load(env(map[string]string{
		"LOG_LEVEL":        "loud",
		"SHUTDOWN_TIMEOUT": "soon",
		"RATE_LIMIT":       "-1",
	}))
	fmt.Println(err)
	// Output:
	// <nil>
	// :8080 https://sho.rt DEBUG 30s 10
	// BASE_URL is required
	// LOG_LEVEL: slog: level string "loud": unknown name
	// SHUTDOWN_TIMEOUT must be a positive duration like 15s, got "soon"
	// RATE_LIMIT must be a positive number, got "-1"
}

// [end example]

func TestSecretFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db-url")
	os.WriteFile(path, []byte("postgres://user:hunter2@db/app\n"), 0o600)

	cfg, err := Load(env(map[string]string{"BASE_URL": "https://sho.rt", "DATABASE_URL_FILE": path}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DatabaseURL != "postgres://user:hunter2@db/app" {
		t.Fatalf("DatabaseURL = %q", cfg.DatabaseURL)
	}

	var buf bytes.Buffer
	slog.New(slog.NewTextHandler(&buf, nil)).Info("starting", "config", cfg)
	if strings.Contains(buf.String(), "hunter2") {
		t.Fatalf("secret leaked into logs: %s", buf.String())
	}
	t.Log(buf.String())
}

func TestSecretBothSet(t *testing.T) {
	_, err := Load(env(map[string]string{"BASE_URL": "https://sho.rt", "DATABASE_URL": "x", "DATABASE_URL_FILE": "/y"}))
	if err == nil || !strings.Contains(err.Error(), "not both") {
		t.Fatalf("err = %v", err)
	}
}
