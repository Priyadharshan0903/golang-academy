# Build the greet CLI with version information stamped in.
#
#   make build       build for this machine into ./bin
#   make release     cross-compile for Linux, macOS and Windows into ./dist
#   make test        run the tests

PKG      := github.com/Priyadharshan0903/gopath-academy/projects/greet
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE     ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X $(PKG)/internal/version.Version=$(VERSION) \
	-X $(PKG)/internal/version.Commit=$(COMMIT) \
	-X $(PKG)/internal/version.Date=$(DATE)

PLATFORMS := linux/amd64 linux/arm64 darwin/arm64 windows/amd64

.PHONY: build release test clean

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/greet ./cmd/greet

release:
	@for p in $(PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; ext=; \
		[ "$$os" = windows ] && ext=.exe; \
		echo "building $$os/$$arch"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
			go build -trimpath -ldflags "$(LDFLAGS)" \
			-o dist/greet-$$os-$$arch$$ext ./cmd/greet || exit 1; \
	done

test:
	go test ./...

clean:
	rm -rf bin dist
