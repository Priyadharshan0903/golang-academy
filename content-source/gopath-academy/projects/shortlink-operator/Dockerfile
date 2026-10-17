# syntax=docker/dockerfile:1
# Container image for the ShortLink operator (lesson 15.4).
#   go mod tidy && docker build -t shortlink-operator:dev .
FROM golang:1.27 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/manager ./cmd/manager

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/manager /manager
# A numeric user, so Kubernetes can verify runAsNonRoot (lesson 15.2).
USER 65532:65532
ENTRYPOINT ["/manager"]
