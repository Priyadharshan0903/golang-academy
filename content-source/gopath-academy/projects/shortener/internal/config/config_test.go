package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func TestDefaults(t *testing.T) {
	c, err := Load(env(nil))
	if err != nil {
		t.Fatal(err)
	}
	if c.Addr != ":8080" || c.AdminAddr != ":9090" || c.BaseURL.String() != "http://localhost:8080" ||
		c.RateLimit != 5 || c.RateBurst != 10 || c.ShutdownTimeout != 10*time.Second || c.ShutdownDelay != 0 ||
		c.DatabaseURL != "" || !c.Migrate || c.RedisAddr != "" || c.CacheTTL != time.Hour {
		t.Fatalf("unexpected defaults: %+v", c)
	}
}

func TestAllErrorsReported(t *testing.T) {
	_, err := Load(env(map[string]string{
		"BASE_URL": "not a url", "LOG_LEVEL": "loud", "LOG_FORMAT": "xml",
		"RATE_LIMIT": "0", "RATE_BURST": "x", "SHUTDOWN_DELAY": "-1s",
	}))
	if err == nil {
		t.Fatal("expected errors")
	}
	for _, want := range []string{"BASE_URL", "LOG_LEVEL", "LOG_FORMAT", "RATE_LIMIT", "RATE_BURST", "SHUTDOWN_DELAY"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error does not mention %s:\n%v", want, err)
		}
	}
}

func TestDatabaseURLFromFile(t *testing.T) {
	path := t.TempDir() + "/db"
	os.WriteFile(path, []byte("postgres://u:secret@db/shortener\n"), 0o600)
	c, err := Load(env(map[string]string{"DATABASE_URL_FILE": path, "MIGRATE": "false"}))
	if err != nil || c.DatabaseURL != "postgres://u:secret@db/shortener" || c.Migrate {
		t.Fatalf("%+v, %v", c, err)
	}
}
