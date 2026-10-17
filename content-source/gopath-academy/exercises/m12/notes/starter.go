//go:build !solution

package notes

import "net/http"

// NewHandler returns the HTTP API for store:
//
//	GET    /notes       200, a JSON array of all notes ([] when empty)
//	POST   /notes       201, the created note as JSON, and a Location
//	                    header /notes/{id}. 400 if the body is not a
//	                    JSON object, 422 if "title" is empty.
//	GET    /notes/{id}  200 and the note, 404 if it does not exist,
//	                    400 if id is not a number
//	DELETE /notes/{id}  204 with no body, 404 if it does not exist
//
// Successful responses have Content-Type application/json (except 204).
// Other methods must get 405 Method Not Allowed: ServeMux does that for
// you if you put methods in the patterns.
func NewHandler(store *Store) http.Handler {
	mux := http.NewServeMux()
	// TODO: register the four routes.
	return mux
}
