package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/link"
)

// problem is an RFC 9457 problem details body.
type problem struct {
	Title     string `json:"title"`
	Status    int    `json:"status"`
	Detail    string `json:"detail,omitzero"`
	RequestID string `json:"request_id,omitzero"`
}

// [start errors]

// statusFor maps an error to an HTTP status code. This is the one place
// where domain errors meet HTTP, as planned in milestone 1.
func statusFor(err error) int {
	var re *requestError
	switch {
	case errors.As(err, &re):
		return re.status
	case errors.Is(err, link.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, link.ErrCodeTaken):
		return http.StatusConflict
	case errors.Is(err, link.ErrInvalidURL), errors.Is(err, link.ErrInvalidCode):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

// writeError writes err as a problem response. Details of client errors
// are shown to the client; server errors are logged, and the client
// gets only a generic message and the request ID to quote.
func (a *api) writeError(w http.ResponseWriter, r *http.Request, err error) {
	status := statusFor(err)
	p := problem{Title: http.StatusText(status), Status: status, RequestID: requestIDFrom(r.Context())}
	if status < 500 {
		p.Detail = err.Error()
	} else {
		a.Logger.LogAttrs(r.Context(), slog.LevelError, "request failed",
			slog.String("error", err.Error()), slog.String("request_id", p.RequestID))
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(p)
}

// [end errors]
