package jsonapi

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

func post(body, contentType string) (int, string) {
	req := httptest.NewRequest("POST", "/users", strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	HandleCreateUser(rec, req)
	return rec.Code, strings.TrimSpace(rec.Body.String())
}

// [start example]

func Example() {
	for _, body := range []string{
		`{"name":"Ada","email":"ada@example.com","age":36}`,
		`{"name":"Bob","email":"bob@example.com"}`,
		`{"name":"","email":"nope","age":-1}`,
		`{"name":"Ada","email":"ada@example.com","age":"36"}`,
		`{"name":"Ada"`,
		`{"name":"Ada","email":"ada@example.com"} {"name":"Eve"}`,
		``,
	} {
		code, resp := post(body, "application/json")
		fmt.Println(code, resp)
	}
	code, resp := post(`{}`, "text/plain")
	fmt.Println(code, resp)
	// Output:
	// 201 {"id":1,"name":"Ada","email":"ada@example.com","age":36}
	// 201 {"id":1,"name":"Bob","email":"bob@example.com"}
	// 422 {"title":"Unprocessable Entity","status":422,"detail":"validation failed","errors":{"age":"must be between 0 and 150","email":"must be a valid email address","name":"is required"}}
	// 400 {"title":"Bad Request","status":400,"detail":"field \"age\" has the wrong type"}
	// 400 {"title":"Bad Request","status":400,"detail":"body is not valid JSON"}
	// 400 {"title":"Bad Request","status":400,"detail":"body must contain a single JSON object"}
	// 400 {"title":"Bad Request","status":400,"detail":"body must not be empty"}
	// 415 {"title":"Unsupported Media Type","status":415,"detail":"Content-Type must be application/json"}
}

// [end example]

func TestUnknownField(t *testing.T) {
	code, resp := post(`{"name":"Ada","email":"ada@example.com","colour":"blue"}`, "application/json")
	t.Logf("%d %s", code, resp)
	// Only check what is stable: the wording of this error depends on the Go version.
	if code != 400 || !strings.Contains(resp, "colour") {
		t.Fatalf("got %d %s", code, resp)
	}
}

func TestTooLarge(t *testing.T) {
	body := `{"name":"` + strings.Repeat("a", 2<<20) + `","email":"a@b.c"}`
	code, resp := post(body, "application/json; charset=utf-8")
	if code != 413 {
		t.Fatalf("got %d %s", code, resp)
	}
}

func TestLocationHeader(t *testing.T) {
	req := httptest.NewRequest("POST", "/users", strings.NewReader(`{"name":"Ada","email":"ada@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	HandleCreateUser(rec, req)
	if rec.Header().Get("Location") != "/users/1" {
		t.Fatalf("Location = %q", rec.Header().Get("Location"))
	}
}
