//go:build solution

// Package apiclient practices net/http and encoding/json by building a
// small, correct client for a JSON API.
package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ErrNotFound is returned when the API responds with 404.
var ErrNotFound = errors.New("not found")

// APIError is returned for any other non-2xx response.
type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string { return fmt.Sprintf("api error %d: %s", e.Status, e.Message) }

// User is the API's user resource.
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Client talks to the API at BaseURL.
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

// GetUser fetches a user by ID.
func (c *Client) GetUser(ctx context.Context, id int) (User, error) {
	var u User
	err := c.do(ctx, http.MethodGet, fmt.Sprintf("/users/%d", id), nil, http.StatusOK, &u)
	if err != nil {
		return User{}, fmt.Errorf("get user %d: %w", id, err)
	}
	return u, nil
}

// CreateUser creates a user and returns it.
func (c *Client) CreateUser(ctx context.Context, name, email string) (User, error) {
	var u User
	body := map[string]string{"name": name, "email": email}
	if err := c.do(ctx, http.MethodPost, "/users", body, http.StatusCreated, &u); err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// do performs one request and decodes a JSON response into out.
func (c *Client) do(ctx context.Context, method, path string, in any, want int, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	limited := io.LimitReader(resp.Body, 1<<20)

	switch {
	case resp.StatusCode == want:
		return json.NewDecoder(limited).Decode(out)
	case resp.StatusCode == http.StatusNotFound:
		return ErrNotFound
	default:
		var e struct {
			Error string `json:"error"`
		}
		json.NewDecoder(limited).Decode(&e) // best effort
		return &APIError{Status: resp.StatusCode, Message: e.Error}
	}
}
