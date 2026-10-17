# syntax=docker/dockerfile:1
#
# Container image for the shortener (lesson 15.1). From this directory:
#
#   go mod tidy                       # once, so go.sum exists
#   docker build -t shortener:dev .
#   docker run --rm -p 8080:8080 shortener:dev

# ---------- build stage: the full Go toolchain ----------
FROM golang:1.27 AS build
WORKDIR /src

# Download dependencies in their own layer: it is rebuilt only when
# go.mod or go.sum change, not on every code change.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
# TARGETOS and TARGETARCH are set by docker buildx for each platform of
# a multi-platform build (--platform linux/amd64,linux/arm64). Go
# cross-compiles, so the build stage never needs emulation.
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/shortener ./cmd/shortener

# ---------- runtime stage: just the binary ----------
# distroless/static has CA certificates, time zone data, /etc/passwd
# with a nonroot user, and nothing else: no shell, no package manager.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/shortener /shortener
# A numeric user, so Kubernetes can verify runAsNonRoot (lesson 15.2).
USER 65532:65532
EXPOSE 8080 9090
ENTRYPOINT ["/shortener"]
