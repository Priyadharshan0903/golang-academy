# syntax=docker/dockerfile:1
#
# Build from the repository root:
#   docker build -f deploy/Dockerfile -t gopath-academy:dev .
#   docker run --rm -p 8080:8080 gopath-academy:dev

ARG GO_VERSION=1.24
ARG HUGO_VERSION=0.168.0

# ---------- 1. Build the static site with Hugo ----------
FROM golang:${GO_VERSION}-bookworm AS site
ARG HUGO_VERSION
ARG TARGETARCH
RUN curl -fsSL -o /tmp/hugo.deb \
      "https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_extended_${HUGO_VERSION}_linux-${TARGETARCH:-amd64}.deb" \
 && dpkg -i /tmp/hugo.deb \
 && rm /tmp/hugo.deb
WORKDIR /src
# Hugo needs the code folders too: lessons embed files from them.
COPY site/ site/
COPY examples/ examples/
COPY exercises/ exercises/
COPY projects/ projects/
RUN cd site && hugo --minify --gc

# ---------- 2. Build the file server ----------
FROM golang:${GO_VERSION}-bookworm AS server
WORKDIR /src
COPY go.mod ./
COPY deploy/server/ deploy/server/
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/server ./deploy/server

# ---------- 3. Minimal runtime image ----------
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=server /out/server /server
COPY --from=site /src/site/public /site
EXPOSE 8080
# A numeric user, so Kubernetes can verify runAsNonRoot (lesson 15.2).
USER 65532:65532
ENTRYPOINT ["/server", "-addr", ":8080", "-dir", "/site"]
