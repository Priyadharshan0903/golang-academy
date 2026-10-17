package link

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func newTestService() *Service {
	s := NewService(NewMemoryStore())
	s.now = func() time.Time { return time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC) }
	return s
}

func TestShortenAndResolve(t *testing.T) {
	ctx := context.Background()
	s := newTestService()

	l, err := s.Shorten(ctx, "https://go.dev/doc/", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Code) != generatedLength {
		t.Errorf("generated code %q has length %d", l.Code, len(l.Code))
	}
	if !l.CreatedAt.Equal(s.now()) {
		t.Errorf("CreatedAt = %v", l.CreatedAt)
	}

	got, err := s.Resolve(ctx, l.Code)
	if err != nil || got.URL != "https://go.dev/doc/" {
		t.Errorf("Resolve = %+v, %v", got, err)
	}
}

func TestCustomCode(t *testing.T) {
	ctx := context.Background()
	s := newTestService()

	if _, err := s.Shorten(ctx, "https://example.com", "my-link"); err != nil {
		t.Fatal(err)
	}
	_, err := s.Shorten(ctx, "https://other.example", "my-link")
	if !errors.Is(err, ErrCodeTaken) {
		t.Errorf("second Shorten with same code: %v, want ErrCodeTaken", err)
	}
	if !strings.Contains(err.Error(), "my-link") {
		t.Errorf("error should mention the code: %v", err)
	}
}

func TestValidation(t *testing.T) {
	ctx := context.Background()
	s := newTestService()

	badURLs := []string{"", "ftp://example.com", "example.com", "https://", "http://bad host"}
	for _, u := range badURLs {
		if _, err := s.Shorten(ctx, u, ""); !errors.Is(err, ErrInvalidURL) {
			t.Errorf("Shorten(%q) error = %v, want ErrInvalidURL", u, err)
		}
	}
	badCodes := []string{"ab", strings.Repeat("x", 33), "has space", "emoji🙂"}
	for _, c := range badCodes {
		if _, err := s.Shorten(ctx, "https://example.com", c); !errors.Is(err, ErrInvalidCode) {
			t.Errorf("Shorten(code %q) error = %v, want ErrInvalidCode", c, err)
		}
	}
}

func TestResolveMissing(t *testing.T) {
	_, err := newTestService().Resolve(context.Background(), "nope123")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Resolve missing = %v, want ErrNotFound", err)
	}
}

func TestRetriesOnCollision(t *testing.T) {
	ctx := context.Background()
	s := newTestService()
	codes := []string{"taken01", "taken01", "fresh01"}
	s.rand = func(int) string {
		c := codes[0]
		codes = codes[1:]
		return c
	}
	if _, err := s.Shorten(ctx, "https://a.example", ""); err != nil {
		t.Fatal(err)
	}
	l, err := s.Shorten(ctx, "https://b.example", "")
	if err != nil || l.Code != "fresh01" {
		t.Errorf("after a collision got %+v, %v; want code fresh01", l, err)
	}
}

func TestGivesUpAfterTooManyCollisions(t *testing.T) {
	ctx := context.Background()
	s := newTestService()
	s.rand = func(int) string { return "always1" }
	s.Shorten(ctx, "https://a.example", "")

	_, err := s.Shorten(ctx, "https://b.example", "")
	if !errors.Is(err, ErrCodeTaken) || !strings.Contains(err.Error(), "attempts") {
		t.Errorf("got %v, want an error mentioning attempts and wrapping ErrCodeTaken", err)
	}
}

func TestRandomCodeAlphabet(t *testing.T) {
	for range 100 {
		c := randomCode(16)
		if strings.ContainsAny(c, "0O1lI") || len(c) != 16 {
			t.Fatalf("bad code %q", c)
		}
	}
}
