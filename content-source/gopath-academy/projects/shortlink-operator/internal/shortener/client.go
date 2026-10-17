// Package shortener is a small client for the shortener's HTTP API,
// with exactly the calls the operator needs.
package shortener

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrNotFound = errors.New("link not found")
	ErrConflict = errors.New("code already in use")
)

// Link is the API's representation of a short link.
type Link struct {
	Code     string `json:"code"`
	URL      string `json:"url"`
	ShortURL string `json:"short_url"`
}

// Client calls the API at BaseURL.
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

// New returns a client with a sensible timeout.
func New(baseURL string) *Client {
	return &Client{BaseURL: strings.TrimSuffix(baseURL, "/"), HTTP: &http.Client{Timeout: 10 * time.Second}}
}

// Get returns the link for code, or ErrNotFound.
func (c *Client) Get(ctx context.Context, code string) (Link, error) {
	var l Link
	err := c.do(ctx, http.MethodGet, "/api/links/"+url.PathEscape(code), nil, http.StatusOK, &l)
	return l, err
}

// Create creates a link, or returns ErrConflict if the code is taken.
func (c *Client) Create(ctx context.Context, code, target string) (Link, error) {
	var l Link
	body, _ := json.Marshal(map[string]string{"code": code, "url": target})
	err := c.do(ctx, http.MethodPost, "/api/links", body, http.StatusCreated, &l)
	return l, err
}

// Delete deletes the link for code, or returns ErrNotFound.
func (c *Client) Delete(ctx context.Context, code string) error {
	return c.do(ctx, http.MethodDelete, "/api/links/"+url.PathEscape(code), nil, http.StatusNoContent, nil)
}

func (c *Client) do(ctx context.Context, method, path string, body []byte, want int, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	switch {
	case resp.StatusCode == want:
		if out != nil {
			return json.Unmarshal(data, out)
		}
		return nil
	case resp.StatusCode == http.StatusNotFound:
		return ErrNotFound
	case resp.StatusCode == http.StatusConflict:
		return ErrConflict
	default:
		var p struct{ Detail string }
		json.Unmarshal(data, &p)
		return fmt.Errorf("%s %s: HTTP %d %s", method, path, resp.StatusCode, p.Detail)
	}
}
