//go:build !solution

// Package apiclient practices net/http and encoding/json by building a
// small, correct client for a JSON API.
package apiclient

import (
	"context"
	"errors"
	"net/http"
)

// ErrNotFound is returned when the API responds with 404.
var ErrNotFound = errors.New("not found")

// APIError is returned for any other non-2xx response. The API sends
// error bodies like {"error": "message"}.
type APIError struct {
	Status  int    // HTTP status code
	Message string // the "error" field of the body, if present
}

func (e *APIError) Error() string {
	// TODO: implement me. Suggested: "api error 503: maintenance".
	return ""
}

// User is the API's user resource:
//
//	{"id": 7, "name": "Ana", "email": "ana@example.com"}
type User struct {
	ID    int
	Name  string
	Email string
}

// Client talks to the API at BaseURL using HTTP. HTTP must not be nil.
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

// GetUser fetches GET {BaseURL}/users/{id}.
//
//   - It sends "Accept: application/json".
//   - It respects ctx (use http.NewRequestWithContext).
//   - 200: decode the body into a User.
//   - 404: return an error wrapping ErrNotFound.
//   - Other statuses: return an *APIError (wrapped is fine).
//   - It always closes the response body.
func (c *Client) GetUser(ctx context.Context, id int) (User, error) {
	// TODO: implement me.
	return User{}, nil
}

// CreateUser sends POST {BaseURL}/users with a JSON body
// {"name": ..., "email": ...} and "Content-Type: application/json",
// and returns the created User from a 201 response. Other statuses
// are handled like in GetUser.
func (c *Client) CreateUser(ctx context.Context, name, email string) (User, error) {
	// TODO: implement me.
	return User{}, nil
}
