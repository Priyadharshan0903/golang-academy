//go:build solution

package notes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// NewHandler returns the HTTP API for store. See the starter for the
// specification.
func NewHandler(store *Store) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /notes", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, store.List())
	})

	mux.HandleFunc("POST /notes", func(w http.ResponseWriter, r *http.Request) {
		var in Note
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		if err := dec.Decode(&in); err != nil {
			http.Error(w, "body must be a JSON object", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(in.Title) == "" {
			http.Error(w, "title is required", http.StatusUnprocessableEntity)
			return
		}
		n := store.Create(Note{Title: in.Title, Body: in.Body}) // ignore any client-sent ID
		w.Header().Set("Location", fmt.Sprintf("/notes/%d", n.ID))
		writeJSON(w, http.StatusCreated, n)
	})

	mux.HandleFunc("GET /notes/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseID(w, r)
		if !ok {
			return
		}
		n, found := store.Get(id)
		if !found {
			http.Error(w, "note not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, n)
	})

	mux.HandleFunc("DELETE /notes/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseID(w, r)
		if !ok {
			return
		}
		if !store.Delete(id) {
			http.Error(w, "note not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	return mux
}

func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "id must be a number", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
