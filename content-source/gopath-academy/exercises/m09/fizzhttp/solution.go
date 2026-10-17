//go:build solution

// Package fizzhttp practices writing an HTTP handler that is easy to
// test with net/http/httptest.
package fizzhttp

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// Handler serves GET /fizzbuzz?n=N with a JSON FizzBuzz sequence.
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	n, err := strconv.Atoi(r.URL.Query().Get("n"))
	if err != nil || n < 1 || n > 1000 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "n must be an integer from 1 to 1000"})
		return
	}
	result := make([]string, n)
	for i := 1; i <= n; i++ {
		switch {
		case i%15 == 0:
			result[i-1] = "FizzBuzz"
		case i%3 == 0:
			result[i-1] = "Fizz"
		case i%5 == 0:
			result[i-1] = "Buzz"
		default:
			result[i-1] = strconv.Itoa(i)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"n": n, "result": result})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
