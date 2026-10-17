// Package jsonapi holds the examples for lesson 12.3.
package jsonapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/mail"
	"strings"
)

// [start problem]

// Problem is an error response in the "problem details" format of
// RFC 9457, served as application/problem+json.
type Problem struct {
	Title  string            `json:"title"`
	Status int               `json:"status"`
	Detail string            `json:"detail,omitempty"`
	Errors map[string]string `json:"errors,omitempty"` // our extension: errors per field
}

// WriteJSON writes v as JSON with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// WriteProblem writes p as a problem details response.
func WriteProblem(w http.ResponseWriter, p Problem) {
	if p.Title == "" {
		p.Title = http.StatusText(p.Status)
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	json.NewEncoder(w).Encode(p)
}

// [end problem]

// [start decode]

// A RequestError is a problem with the request, reported to the client.
type RequestError struct {
	Status int
	Detail string
}

func (e *RequestError) Error() string { return e.Detail }

// DecodeJSON decodes a request body holding exactly one JSON object
// into a T. It limits the body to maxBytes and rejects unknown fields.
// Every error it returns is a *RequestError that is safe to show.
func DecodeJSON[T any](w http.ResponseWriter, r *http.Request, maxBytes int64) (T, error) {
	var v T
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "application/json" {
		return v, &RequestError{http.StatusUnsupportedMediaType, "Content-Type must be application/json"}
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&v); err != nil {
		return v, decodeError(err)
	}
	// Exactly one value: anything after it, even a second object, is an error.
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return v, &RequestError{http.StatusBadRequest, "body must contain a single JSON object"}
	}
	return v, nil
}

func decodeError(err error) *RequestError {
	var (
		syntaxErr *json.SyntaxError
		typeErr   *json.UnmarshalTypeError
		tooBig    *http.MaxBytesError
	)
	switch {
	case errors.As(err, &tooBig):
		return &RequestError{http.StatusRequestEntityTooLarge, fmt.Sprintf("body must be at most %d bytes", tooBig.Limit)}
	case errors.Is(err, io.EOF):
		return &RequestError{http.StatusBadRequest, "body must not be empty"}
	case errors.As(err, &syntaxErr), errors.Is(err, io.ErrUnexpectedEOF):
		return &RequestError{http.StatusBadRequest, "body is not valid JSON"}
	case errors.As(err, &typeErr) && typeErr.Field != "":
		return &RequestError{http.StatusBadRequest, fmt.Sprintf("field %q has the wrong type", typeErr.Field)}
	default:
		// For example an unknown field. The exact wording comes from
		// encoding/json and differs between Go versions.
		return &RequestError{http.StatusBadRequest, strings.TrimPrefix(err.Error(), "json: ")}
	}
}

// [end decode]

// [start handler]

// CreateUser is the body of POST /users.
type CreateUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// Validate checks the business rules, returning one message per field.
func (c CreateUser) Validate() map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(c.Name) == "" {
		errs["name"] = "is required"
	}
	if _, err := mail.ParseAddress(c.Email); err != nil {
		errs["email"] = "must be a valid email address"
	}
	if c.Age < 0 || c.Age > 150 {
		errs["age"] = "must be between 0 and 150"
	}
	return errs
}

// User is what the API returns.
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age,omitzero"` // omitted when 0 (Go 1.24+)
}

// HandleCreateUser decodes, validates, and responds.
func HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	in, err := DecodeJSON[CreateUser](w, r, 1<<20) // 1 MiB is plenty for this body
	if err != nil {
		var re *RequestError
		errors.As(err, &re)
		WriteProblem(w, Problem{Status: re.Status, Detail: re.Detail})
		return
	}
	if errs := in.Validate(); len(errs) > 0 {
		WriteProblem(w, Problem{Status: http.StatusUnprocessableEntity, Detail: "validation failed", Errors: errs})
		return
	}
	user := User{ID: 1, Name: in.Name, Email: in.Email, Age: in.Age}
	w.Header().Set("Location", fmt.Sprintf("/users/%d", user.ID))
	WriteJSON(w, http.StatusCreated, user)
}

// [end handler]
