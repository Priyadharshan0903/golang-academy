package layered

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"testing"
)

// fakeEnv has one config file and the given environment variables.
func fakeEnv(vars map[string]string, files map[string]string) Env {
	return Env{
		Getenv: func(k string) string { return vars[k] },
		ReadFile: func(name string) ([]byte, error) {
			if s, ok := files[name]; ok {
				return []byte(s), nil
			}
			return nil, fs.ErrNotExist
		},
		ConfigDir: func() (string, error) { return "/home/ada/.config", nil },
	}
}

// [start example]

func ExampleLoad() {
	files := map[string]string{
		"/home/ada/.config/shortctl/config.json": `{"server": "https://sho.rt", "timeout": "3s"}`,
	}
	env := fakeEnv(map[string]string{"SHORTCTL_TIMEOUT": "5s"}, files)

	cfg, args, err := Load([]string{"-output", "json", "create", "https://go.dev"}, env)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("server  = %-14s from %s\n", cfg.Server, cfg.Source["server"])
	fmt.Printf("timeout = %-14s from %s\n", cfg.Timeout, cfg.Source["timeout"])
	fmt.Printf("output  = %-14s from %s\n", cfg.Output, cfg.Source["output"])
	fmt.Println("args   =", args)
	// Output:
	// server  = https://sho.rt from file /home/ada/.config/shortctl/config.json
	// timeout = 5s             from env SHORTCTL_TIMEOUT
	// output  = json           from flag -output
	// args   = [create https://go.dev]
}

// [end example]

func TestDefaultsWithoutFile(t *testing.T) {
	cfg, _, err := Load(nil, fakeEnv(nil, nil))
	if err != nil || cfg.Server != "http://localhost:8080" || cfg.Source["server"] != "default" {
		t.Fatalf("%+v %v", cfg, err)
	}
}

func TestExplicitFileMustExist(t *testing.T) {
	_, _, err := Load([]string{"-config", "/etc/nope.json"}, fakeEnv(nil, nil))
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("err = %v", err)
	}
}

func TestFileTypo(t *testing.T) {
	env := fakeEnv(map[string]string{"SHORTCTL_CONFIG": "/c.json"}, map[string]string{"/c.json": `{"sever": "x"}`})
	_, _, err := Load(nil, env)
	if err == nil || !strings.Contains(err.Error(), "sever") {
		t.Fatalf("err = %v", err)
	}
}

func TestFlagDefaultDoesNotOverride(t *testing.T) {
	env := fakeEnv(map[string]string{"SHORTCTL_SERVER": "https://env.example"}, nil)
	cfg, _, _ := Load([]string{"-timeout", "1s"}, env)
	if cfg.Server != "https://env.example" {
		t.Fatalf("server = %q: an unset -server flag overrode the environment", cfg.Server)
	}
}

func TestInvalidOutputNamesSource(t *testing.T) {
	_, _, err := Load(nil, fakeEnv(map[string]string{"SHORTCTL_OUTPUT": "yaml"}, nil))
	if err == nil || !strings.Contains(err.Error(), "env SHORTCTL_OUTPUT") {
		t.Fatalf("err = %v", err)
	}
}
