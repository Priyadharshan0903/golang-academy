// Package layered holds the configuration example for lesson 14.2:
// settings from defaults, a config file, environment variables and
// flags, with a clear order of precedence.
package layered

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// [start config]

// Config is the effective configuration of a CLI called shortctl.
type Config struct {
	Server  string        // base URL of the API
	Timeout time.Duration // per request
	Output  string        // "text" or "json"
	Source  map[string]string
}

// fileConfig is the JSON config file. Pointers tell "not set" apart
// from "set to the zero value".
type fileConfig struct {
	Server  *string `json:"server"`
	Timeout *string `json:"timeout"`
	Output  *string `json:"output"`
}

// Env abstracts the outside world so tests can control it.
type Env struct {
	Getenv    func(string) string
	ReadFile  func(string) ([]byte, error)
	ConfigDir func() (string, error) // os.UserConfigDir in real use
}

// OS is the real environment.
var OS = Env{Getenv: os.Getenv, ReadFile: os.ReadFile, ConfigDir: os.UserConfigDir}

// Load resolves the configuration. Later sources override earlier ones:
//
//	defaults < config file < environment variables < flags
//
// It returns the remaining positional arguments.
func Load(args []string, env Env) (Config, []string, error) {
	cfg := Config{Server: "http://localhost:8080", Timeout: 10 * time.Second, Output: "text",
		Source: map[string]string{"server": "default", "timeout": "default", "output": "default"}}

	fs := flag.NewFlagSet("shortctl", flag.ContinueOnError)
	configPath := fs.String("config", "", "config file (default: $SHORTCTL_CONFIG or <user config dir>/shortctl/config.json)")
	server := fs.String("server", "", "API base URL")
	timeout := fs.Duration("timeout", 0, "request timeout")
	output := fs.String("output", "", "output format: text or json")
	if err := fs.Parse(args); err != nil {
		return cfg, nil, err
	}

	// 1. The config file, if there is one.
	if err := cfg.applyFile(*configPath, env); err != nil {
		return cfg, nil, err
	}

	// 2. Environment variables.
	for name, apply := range map[string]func(string) error{
		"SHORTCTL_SERVER":  func(v string) error { cfg.Server = v; return nil },
		"SHORTCTL_TIMEOUT": func(v string) (err error) { cfg.Timeout, err = time.ParseDuration(v); return },
		"SHORTCTL_OUTPUT":  func(v string) error { cfg.Output = v; return nil },
	} {
		if v := env.Getenv(name); v != "" {
			if err := apply(v); err != nil {
				return cfg, nil, fmt.Errorf("%s: %w", name, err)
			}
			cfg.Source[keyOf(name)] = "env " + name
		}
	}

	// 3. Flags, but only those given on the command line. fs.Visit
	// visits exactly the flags that were set, so a flag's default
	// never overrides the file or the environment.
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "server":
			cfg.Server = *server
		case "timeout":
			cfg.Timeout = *timeout
		case "output":
			cfg.Output = *output
		default:
			return
		}
		cfg.Source[f.Name] = "flag -" + f.Name
	})

	if cfg.Output != "text" && cfg.Output != "json" {
		return cfg, nil, fmt.Errorf("output must be text or json, got %q (from %s)", cfg.Output, cfg.Source["output"])
	}
	return cfg, fs.Args(), nil
}

// [end config]

// [start file]

func (cfg *Config) applyFile(path string, env Env) error {
	explicit := path != ""
	if !explicit {
		path = env.Getenv("SHORTCTL_CONFIG")
		explicit = path != ""
	}
	if !explicit {
		dir, err := env.ConfigDir()
		if err != nil {
			return nil // no config directory on this system: skip the file
		}
		path = filepath.Join(dir, "shortctl", "config.json")
	}

	b, err := env.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) && !explicit {
		return nil // the default file is optional; an explicit one is not
	}
	if err != nil {
		return err
	}

	var fc fileConfig
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields() // a typo in a key should be an error, not ignored
	if err := dec.Decode(&fc); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	src := "file " + path
	if fc.Server != nil {
		cfg.Server, cfg.Source["server"] = *fc.Server, src
	}
	if fc.Timeout != nil {
		d, err := time.ParseDuration(*fc.Timeout)
		if err != nil {
			return fmt.Errorf("%s: timeout: %w", path, err)
		}
		cfg.Timeout, cfg.Source["timeout"] = d, src
	}
	if fc.Output != nil {
		cfg.Output, cfg.Source["output"] = *fc.Output, src
	}
	return nil
}

// [end file]

func keyOf(env string) string {
	switch env {
	case "SHORTCTL_SERVER":
		return "server"
	case "SHORTCTL_TIMEOUT":
		return "timeout"
	default:
		return "output"
	}
}
