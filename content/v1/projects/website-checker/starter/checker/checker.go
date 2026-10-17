// Package checker performs bounded, cancelable HTTP health checks.
package checker

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const MaxWorkers = 8
const MaxURLs = 1000

type Config struct {
	Workers        int
	RequestTimeout time.Duration
	Client         *http.Client
}

type Result struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	Error      string `json:"error,omitempty"`
}

// Checker may be shared by callers. Do not mutate its client's configuration
// during a run; the client's transport must support concurrent use.
type Checker struct {
	workers int
	timeout time.Duration
	client  *http.Client
}

func New(config Config) (*Checker, error) {
	if config.Workers < 1 || config.Workers > MaxWorkers {
		return nil, fmt.Errorf("workers must be between 1 and %d", MaxWorkers)
	}
	if config.RequestTimeout <= 0 {
		return nil, fmt.Errorf("request timeout must be positive")
	}
	client := config.Client
	if client == nil {
		client = &http.Client{}
	}
	return &Checker{workers: config.Workers, timeout: config.RequestTimeout, client: client}, nil
}

// Check is your implementation task. Keep the signature and Result contract.
// Start with validation, then implement one request, then add bounded workers.
func (c *Checker) Check(ctx context.Context, urls []string) ([]Result, error) {
	return nil, fmt.Errorf("TODO: implement bounded, cancelable checks")
}
