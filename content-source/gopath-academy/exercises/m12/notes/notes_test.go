package notes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCreateAndGet(t *testing.T) {
	h := NewHandler(NewStore())

	rec := do(t, h, "POST", "/notes", `{"title":"groceries","body":"milk"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /notes: status %d, want 201 (body %q)", rec.Code, rec.Body)
	}
	if loc := rec.Header().Get("Location"); loc != "/notes/1" {
		t.Errorf("Location = %q, want /notes/1", loc)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q", ct)
	}
	var created Note
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || created != (Note{1, "groceries", "milk"}) {
		t.Fatalf("created %+v, err %v", created, err)
	}

	rec = do(t, h, "GET", "/notes/1", "")
	var got Note
	if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &got) != nil || got != created {
		t.Fatalf("GET /notes/1: %d %s", rec.Code, rec.Body)
	}
}

func TestList(t *testing.T) {
	h := NewHandler(NewStore())
	rec := do(t, h, "GET", "/notes", "")
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Fatalf("empty list: %d %q, want 200 []", rec.Code, rec.Body)
	}
	do(t, h, "POST", "/notes", `{"title":"a"}`)
	do(t, h, "POST", "/notes", `{"title":"b"}`)
	rec = do(t, h, "GET", "/notes", "")
	var list []Note
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil || len(list) != 2 || list[0].Title != "a" || list[1].Title != "b" {
		t.Fatalf("list: %s", rec.Body)
	}
}

func TestErrors(t *testing.T) {
	h := NewHandler(NewStore())
	tests := []struct {
		method, path, body string
		want               int
	}{
		{"POST", "/notes", `not json`, http.StatusBadRequest},
		{"POST", "/notes", `{"title":"  "}`, http.StatusUnprocessableEntity},
		{"GET", "/notes/99", "", http.StatusNotFound},
		{"GET", "/notes/abc", "", http.StatusBadRequest},
		{"DELETE", "/notes/99", "", http.StatusNotFound},
		{"PATCH", "/notes/1", `{}`, http.StatusMethodNotAllowed},
		{"PUT", "/notes", `{}`, http.StatusMethodNotAllowed},
	}
	for _, tt := range tests {
		if rec := do(t, h, tt.method, tt.path, tt.body); rec.Code != tt.want {
			t.Errorf("%s %s %s: status %d, want %d", tt.method, tt.path, tt.body, rec.Code, tt.want)
		}
	}
}

func TestDelete(t *testing.T) {
	h := NewHandler(NewStore())
	do(t, h, "POST", "/notes", `{"title":"temp"}`)
	rec := do(t, h, "DELETE", "/notes/1", "")
	if rec.Code != http.StatusNoContent || rec.Body.Len() != 0 {
		t.Fatalf("DELETE: %d %q, want 204 and no body", rec.Code, rec.Body)
	}
	if rec := do(t, h, "GET", "/notes/1", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("GET after DELETE: %d, want 404", rec.Code)
	}
}
