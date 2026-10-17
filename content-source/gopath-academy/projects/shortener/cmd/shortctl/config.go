package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type config struct {
	Server  string
	Timeout time.Duration
	Output  string
}

// env is everything shortctl reads from the outside world.
type env struct {
	getenv    func(string) string
	readFile  func(string) ([]byte, error)
	configDir func() (string, error)
}

func defaultEnv() env { return env{os.Getenv, os.ReadFile, os.UserConfigDir} }

// loadConfig applies defaults < config file < environment < flags, and
// returns the arguments after the global flags.
func loadConfig(args []string, e env, stderr io.Writer) (config, []string, error) {
	cfg := config{Server: "http://localhost:8080", Timeout: 10 * time.Second, Output: "text"}

	fs := flag.NewFlagSet("shortctl", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { usage(stderr); fs.PrintDefaults() }
	configPath := fs.String("config", "", "JSON config file (default $SHORTCTL_CONFIG, then <user config dir>/shortctl/config.json)")
	server := fs.String("server", "", "API base URL (default http://localhost:8080)")
	timeout := fs.Duration("timeout", 0, "timeout per request (default 10s)")
	output := fs.String("output", "", "output format: text or json (default text)")
	if err := fs.Parse(args); err != nil {
		return cfg, nil, err
	}

	if err := applyFile(&cfg, *configPath, e); err != nil {
		return cfg, nil, err
	}
	if v := e.getenv("SHORTCTL_SERVER"); v != "" {
		cfg.Server = v
	}
	if v := e.getenv("SHORTCTL_OUTPUT"); v != "" {
		cfg.Output = v
	}
	if v := e.getenv("SHORTCTL_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return cfg, nil, fmt.Errorf("SHORTCTL_TIMEOUT: %w", err)
		}
		cfg.Timeout = d
	}
	fs.Visit(func(f *flag.Flag) { // only flags given on the command line
		switch f.Name {
		case "server":
			cfg.Server = *server
		case "timeout":
			cfg.Timeout = *timeout
		case "output":
			cfg.Output = *output
		}
	})
	if cfg.Output != "text" && cfg.Output != "json" {
		return cfg, nil, fmt.Errorf("output must be text or json, got %q", cfg.Output)
	}
	return cfg, fs.Args(), nil
}

func applyFile(cfg *config, path string, e env) error {
	explicit := path != ""
	if !explicit {
		path = e.getenv("SHORTCTL_CONFIG")
		explicit = path != ""
	}
	if !explicit {
		dir, err := e.configDir()
		if err != nil {
			return nil
		}
		path = filepath.Join(dir, "shortctl", "config.json")
	}
	b, err := e.readFile(path)
	if errors.Is(err, fs.ErrNotExist) && !explicit {
		return nil
	}
	if err != nil {
		return err
	}
	var fc struct {
		Server  *string `json:"server"`
		Timeout *string `json:"timeout"`
		Output  *string `json:"output"`
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&fc); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if fc.Server != nil {
		cfg.Server = *fc.Server
	}
	if fc.Output != nil {
		cfg.Output = *fc.Output
	}
	if fc.Timeout != nil {
		d, err := time.ParseDuration(*fc.Timeout)
		if err != nil {
			return fmt.Errorf("%s: timeout: %w", path, err)
		}
		cfg.Timeout = d
	}
	return nil
}
