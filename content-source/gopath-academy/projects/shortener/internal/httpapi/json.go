package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
)

// requestError is a problem with the request itself.
type requestError struct {
	status int
	msg    string
}

func (e *requestError) Error() string { return e.msg }

// decodeJSON decodes a body holding exactly one JSON object (lesson 12.3).
func decodeJSON[T any](w http.ResponseWriter, r *http.Request, maxBytes int64) (T, error) {
	var v T
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "application/json" {
		return v, &requestError{http.StatusUnsupportedMediaType, "Content-Type must be application/json"}
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		var tooBig *http.MaxBytesError
		switch {
		case errors.As(err, &tooBig):
			return v, &requestError{http.StatusRequestEntityTooLarge, fmt.Sprintf("body must be at most %d bytes", tooBig.Limit)}
		case errors.Is(err, io.EOF):
			return v, &requestError{http.StatusBadRequest, "body must not be empty"}
		default:
			return v, &requestError{http.StatusBadRequest, "invalid JSON body: " + strings.TrimPrefix(err.Error(), "json: ")}
		}
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return v, &requestError{http.StatusBadRequest, "body must contain a single JSON object"}
	}
	return v, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
