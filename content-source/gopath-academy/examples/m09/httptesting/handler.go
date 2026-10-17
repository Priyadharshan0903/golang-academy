package httptesting

import (
	"encoding/json"
	"net/http"
	"strings"
)

// [start handler]

// GreetHandler responds to GET /greet?name=Ana with JSON.
func GreetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"greeting": "Hello, " + name + "!"})
}

// [end handler]
