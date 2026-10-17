//go:build !solution

// Package fizzhttp practices writing an HTTP handler that is easy to
// test with net/http/httptest.
package fizzhttp

import "net/http"

// Handler serves GET /fizzbuzz?n=N and responds with JSON:
//
//	{"n": 5, "result": ["1","2","Fizz","4","Buzz"]}
//
// Rules:
//   - Only GET is allowed; other methods get 405 with an "Allow: GET" header.
//   - n must be an integer from 1 to 1000; otherwise respond 400 with a
//     JSON body {"error": "..."} (any message).
//   - Successful responses set "Content-Type: application/json".
//   - Each element is "FizzBuzz" for multiples of 15, "Fizz" for 3,
//     "Buzz" for 5, and the number itself otherwise.
func Handler(w http.ResponseWriter, r *http.Request) {
	// TODO: implement me.
	w.WriteHeader(http.StatusNotImplemented)
}
