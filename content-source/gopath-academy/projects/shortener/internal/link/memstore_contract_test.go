package link_test

import (
	"testing"

	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/link"
	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/link/linktest"
)

func TestMemoryStoreContract(t *testing.T) {
	linktest.Run(t, func(t *testing.T) link.Store { return link.NewMemoryStore() })
}
