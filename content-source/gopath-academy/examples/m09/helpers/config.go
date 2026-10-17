// Package helpers holds the tested examples for lesson 9.2.
package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// LoadEnvFile reads KEY=VALUE lines from path.
func LoadEnvFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("%s:%d: missing '='", filepath.Base(path), i+1)
		}
		out[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return out, nil
}

// Render prints a map as sorted KEY=VALUE lines, prefixed by $APP_ENV.
func Render(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	fmt.Fprintf(&b, "# env: %s\n", os.Getenv("APP_ENV"))
	for _, k := range keys {
		fmt.Fprintf(&b, "%s=%s\n", k, m[k])
	}
	return b.String()
}
