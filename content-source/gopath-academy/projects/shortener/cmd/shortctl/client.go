package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// apiLink mirrors the API's JSON.
type apiLink struct {
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	ShortURL  string    `json:"short_url"`
	CreatedAt time.Time `json:"created_at"`
}

// [start client]

// client talks to the shortener API.
type client struct {
	base string
	http *http.Client
}

// apiError is an error response from the API (RFC 9457 problem details).
type apiError struct {
	Status int    `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

func (e *apiError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s (HTTP %d)", e.Detail, e.Status)
	}
	return fmt.Sprintf("%s (HTTP %d)", e.Title, e.Status)
}

func (c *client) create(ctx context.Context, target, code string) (apiLink, error) {
	body, _ := json.Marshal(map[string]string{"url": target, "code": code})
	return c.do(ctx, http.MethodPost, "/api/links", body, http.StatusCreated)
}

func (c *client) get(ctx context.Context, code string) (apiLink, error) {
	return c.do(ctx, http.MethodGet, "/api/links/"+url.PathEscape(code), nil, http.StatusOK)
}

func (c *client) delete(ctx context.Context, code string) error {
	_, err := c.do(ctx, http.MethodDelete, "/api/links/"+url.PathEscape(code), nil, http.StatusNoContent)
	return err
}

func (c *client) do(ctx context.Context, method, path string, body []byte, want int) (apiLink, error) {
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimSuffix(c.base, "/")+path, bytes.NewReader(body))
	if err != nil {
		return apiLink{}, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return apiLink{}, fmt.Errorf("cannot reach %s: %w", c.base, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return apiLink{}, err
	}

	if resp.StatusCode != want {
		apiErr := &apiError{Status: resp.StatusCode, Title: http.StatusText(resp.StatusCode)}
		json.Unmarshal(data, apiErr) // best effort: keep the defaults if it is not JSON
		return apiLink{}, apiErr
	}
	var l apiLink
	if resp.StatusCode == http.StatusNoContent {
		return l, nil
	}
	if err := json.Unmarshal(data, &l); err != nil {
		return apiLink{}, fmt.Errorf("unexpected response from %s: %w", c.base, err)
	}
	return l, nil
}

// [end client]
