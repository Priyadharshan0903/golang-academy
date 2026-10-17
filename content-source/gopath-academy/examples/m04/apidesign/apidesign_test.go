package apidesign

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// [start consumer]

// The consumer defines the small interface it needs. The producer
// (some storage package) just returns a concrete type.
type userGetter interface {
	GetUser(id int) (string, error)
}

func greeting(store userGetter, id int) string {
	name, err := store.GetUser(id)
	if err != nil {
		return "Hello, stranger"
	}
	return "Hello, " + name
}

// MemoryStore is a concrete type with more methods than greeting needs.
type MemoryStore struct{ users map[int]string }

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{users: map[int]string{}}
}

func (m *MemoryStore) AddUser(id int, name string) { m.users[id] = name }

func (m *MemoryStore) GetUser(id int) (string, error) {
	name, ok := m.users[id]
	if !ok {
		return "", fmt.Errorf("user %d not found", id)
	}
	return name, nil
}

func Example_acceptInterfacesReturnStructs() {
	store := NewMemoryStore() // returns *MemoryStore, not an interface
	store.AddUser(1, "Ana")

	fmt.Println(greeting(store, 1))
	fmt.Println(greeting(store, 2))

	// Output:
	// Hello, Ana
	// Hello, stranger
}

// [end consumer]

// [start options]

type Server struct {
	addr    string
	timeout time.Duration
	logger  io.Writer
}

// Option configures a Server. This is the "functional options" pattern.
type Option func(*Server)

func WithTimeout(d time.Duration) Option { return func(s *Server) { s.timeout = d } }
func WithLogger(w io.Writer) Option      { return func(s *Server) { s.logger = w } }

// NewServer has sensible defaults; options override only what callers care about.
func NewServer(addr string, opts ...Option) *Server {
	s := &Server{addr: addr, timeout: 30 * time.Second, logger: io.Discard}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func Example_functionalOptions() {
	a := NewServer(":8080")
	b := NewServer(":9090", WithTimeout(5*time.Second), WithLogger(os.Stdout))

	fmt.Println(a.addr, a.timeout)
	fmt.Println(b.addr, b.timeout)
	fmt.Fprintln(b.logger, "b logs to stdout")

	// Output:
	// :8080 30s
	// :9090 5s
	// b logs to stdout
}

// [end options]

// [start config]

// For many settings, a plain config struct is often simpler and just
// as clear. The zero value of each field means "use the default".
type ClientConfig struct {
	BaseURL string
	Retries int           // 0 means the default of 3
	Timeout time.Duration // 0 means the default of 10s
}

type Client struct{ cfg ClientConfig }

func NewClient(cfg ClientConfig) *Client {
	if cfg.Retries == 0 {
		cfg.Retries = 3
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &Client{cfg: cfg}
}

func Example_configStruct() {
	c := NewClient(ClientConfig{BaseURL: "https://api.example.com"})
	fmt.Println(c.cfg.Retries, c.cfg.Timeout, strings.HasPrefix(c.cfg.BaseURL, "https"))

	// Output:
	// 3 10s true
}

// [end config]
