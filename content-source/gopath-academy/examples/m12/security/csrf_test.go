//go:build go1.25

package security

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

// [start csrftest]

func ExampleProtectForms() {
	h, _ := ProtectForms(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "done")
	}))
	try := func(method, header, value string) {
		req := httptest.NewRequest(method, "https://shop.example.com/cart", nil)
		if header != "" {
			req.Header.Set(header, value)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		fmt.Printf("%-4s %-30s %d\n", method, header+" "+value, rec.Code)
	}
	try("POST", "Sec-Fetch-Site", "same-origin")
	try("POST", "Sec-Fetch-Site", "cross-site")
	try("GET", "Sec-Fetch-Site", "cross-site")
	try("POST", "Origin", "https://evil.example.net")
	try("POST", "Origin", "https://admin.example.com")
	try("POST", "", "") // curl, or another server: no browser headers
	// Output:
	// POST Sec-Fetch-Site same-origin     200
	// POST Sec-Fetch-Site cross-site      403
	// GET  Sec-Fetch-Site cross-site      200
	// POST Origin https://evil.example.net 403
	// POST Origin https://admin.example.com 200
	// POST                                200
}

// [end csrftest]
