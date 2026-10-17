// Package link is the core of the URL shortener: the Link type, the
// errors callers can check for, and the Service that creates and
// resolves short links on top of any Store.
//
// It has no knowledge of HTTP or databases. Those live in other
// packages and are added in later milestones of the project.
package link

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// Errors that callers can check with errors.Is. The HTTP layer
// (Module 12) maps each of them to a status code.
var (
	ErrNotFound    = errors.New("link not found")
	ErrCodeTaken   = errors.New("code already in use")
	ErrInvalidURL  = errors.New("invalid url")
	ErrInvalidCode = errors.New("invalid code")
)

// Link is a short code pointing at a target URL.
type Link struct {
	Code      string
	URL       string
	CreatedAt time.Time
}

// Store persists links. Implementations must return an error wrapping
// ErrCodeTaken from Create when the code exists, and ErrNotFound from
// Get when it does not.
//
// The context parameter carries cancellation and deadlines; it is
// covered in Module 7. Until then, context.Background() is fine.
type Store interface {
	Create(ctx context.Context, l Link) error
	Get(ctx context.Context, code string) (Link, error)
	// Delete was added in milestone 4, for the Kubernetes operator.
	// It returns an error wrapping ErrNotFound if the code does not exist.
	Delete(ctx context.Context, code string) error
}

// Service implements the shortener's rules on top of a Store.
type Service struct {
	store Store
	now   func() time.Time
	rand  func(n int) string
}

// NewService returns a Service backed by store.
func NewService(store Store) *Service {
	return &Service{store: store, now: time.Now, rand: randomCode}
}

const (
	generatedLength = 7
	maxAttempts     = 5
)

// Shorten creates a link for rawURL. If code is empty, a random code
// is generated; otherwise the requested custom code is used.
func (s *Service) Shorten(ctx context.Context, rawURL, code string) (Link, error) {
	if err := validateURL(rawURL); err != nil {
		return Link{}, err
	}
	if code != "" {
		if err := validateCode(code); err != nil {
			return Link{}, err
		}
		return s.create(ctx, code, rawURL)
	}

	// Random codes can collide. Collisions are rare, so retry a few
	// times and only then give up.
	var err error
	for range maxAttempts {
		var l Link
		l, err = s.create(ctx, s.rand(generatedLength), rawURL)
		if !errors.Is(err, ErrCodeTaken) {
			return l, err
		}
	}
	return Link{}, fmt.Errorf("generate code after %d attempts: %w", maxAttempts, err)
}

func (s *Service) create(ctx context.Context, code, rawURL string) (Link, error) {
	l := Link{Code: code, URL: rawURL, CreatedAt: s.now().UTC()}
	if err := s.store.Create(ctx, l); err != nil {
		return Link{}, fmt.Errorf("create %q: %w", code, err)
	}
	return l, nil
}

// Resolve returns the link for code.
func (s *Service) Resolve(ctx context.Context, code string) (Link, error) {
	if err := validateCode(code); err != nil {
		return Link{}, err
	}
	l, err := s.store.Get(ctx, code)
	if err != nil {
		return Link{}, fmt.Errorf("resolve %q: %w", code, err)
	}
	return l, nil
}

// Delete removes the link for code.
func (s *Service) Delete(ctx context.Context, code string) error {
	if err := validateCode(code); err != nil {
		return err
	}
	if err := s.store.Delete(ctx, code); err != nil {
		return fmt.Errorf("delete %q: %w", code, err)
	}
	return nil
}

func validateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%w: scheme must be http or https, got %q", ErrInvalidURL, u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("%w: missing host", ErrInvalidURL)
	}
	return nil
}

func validateCode(code string) error {
	if len(code) < 3 || len(code) > 32 {
		return fmt.Errorf("%w: length must be 3 to 32, got %d", ErrInvalidCode, len(code))
	}
	for _, r := range code {
		ok := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_'
		if !ok {
			return fmt.Errorf("%w: unexpected character %q", ErrInvalidCode, r)
		}
	}
	return nil
}

const alphabet = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// randomCode returns n characters from an alphabet without look-alike
// characters (no 0/O, 1/l/I), using a cryptographically secure source
// so codes cannot be guessed from earlier ones.
func randomCode(n int) string {
	// Bytes at or above limit are discarded. Without this, the first
	// 256 % len(alphabet) characters would be slightly more likely
	// than the others ("modulo bias").
	const limit = 256 - 256%len(alphabet)
	out := make([]byte, 0, n)
	buf := make([]byte, n*2)
	for len(out) < n {
		rand.Read(buf) // since Go 1.24, Read never returns an error; it crashes the program instead
		for _, b := range buf {
			if int(b) < limit && len(out) < n {
				out = append(out, alphabet[int(b)%len(alphabet)])
			}
		}
	}
	return string(out)
}
