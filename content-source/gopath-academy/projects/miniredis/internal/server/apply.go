package server

import (
	"errors"
	"strings"

	"github.com/Priyadharshan0903/gopath-academy/projects/miniredis/internal/resp"
)

// Apply executes a command without a client connection, for example
// while replaying the append-only file. It returns an error if the
// command produced an error reply.
func (s *Server) Apply(args []string) error {
	var reply strings.Builder
	w := resp.NewWriter(&reply)
	s.execute(args, w)
	w.Flush()
	if strings.HasPrefix(reply.String(), "-") {
		return errors.New(strings.TrimSpace(strings.TrimPrefix(reply.String(), "-")))
	}
	return nil
}
