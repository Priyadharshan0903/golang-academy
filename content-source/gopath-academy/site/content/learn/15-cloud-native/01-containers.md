---
title: "15.1 Containerizing Go"
linkTitle: "15.1 Containerizing Go"
weight: 1
description: "Build small, secure container images for Go services: multi-stage Dockerfiles with build caches, static binaries, distroless versus scratch versus alpine, numeric non-root users, and multi-platform builds."
goVersion: "1.27"
---

## What you will learn

- Why Go is unusually easy to containerize
- A production multi-stage Dockerfile, line by line
- Layer caching and BuildKit cache mounts: a 62 second build becomes 1.6 seconds
- Choosing a base image: distroless, scratch, alpine
- Why `USER 65532` and not `USER nonroot`
- Multi-platform images, and `ko`

## Why Go is easy to containerize

A container image needs everything the program needs at run time. For an interpreted language that means the interpreter, libraries and packages. A Go program built with `CGO_ENABLED=0` is a **single static binary** (lesson 14.4): no runtime, no shared libraries. The image can contain that file and almost nothing else.

That makes images small (fast to pull, fast to start a new pod) and secure (nothing to exploit: no shell, no package manager, no extra libraries with their own vulnerabilities).

## The Dockerfile

This is the shortener's (milestone 4):

{{< example file="projects/shortener/Dockerfile" >}}

Two **stages**. The `build` stage has the full Go toolchain, about 400 MB, and produces the binary. The final stage starts again from a tiny base and copies in only the binary. Nothing from the build stage, not the toolchain and not the source code, ends up in the final image.

Notes on each part:

- **`COPY go.mod go.sum` and `go mod download` come first.** Docker caches each step and reruns a step only if its inputs changed. Dependencies change rarely and code changes constantly, so downloading them in their own step means a code change does not re-download them. (`go.sum` must exist: run `go mod tidy` once and commit it.)
- **`--mount=type=cache`** (BuildKit) keeps the module cache and the **Go build cache** between builds, outside the image. Without it, every build compiles every dependency from scratch.
- **`CGO_ENABLED=0`, `-trimpath`, `-ldflags="-s -w"`**: a static binary, no local paths, no debug symbols (lesson 14.4).
- **`TARGETOS` and `TARGETARCH`**: see multi-platform builds below.
- **`.dockerignore`** keeps `.git`, tests and manifests out of the build context, so they cannot invalidate the cache or leak into the image.

The effect of the cache mounts, measured while writing this lesson:

```text
first build:                         62 s   (compiling pgx, Prometheus client, go-redis, ...)
rebuild after changing one .go file:  1.6 s
```

The image itself:

```text
$ docker history shortener:dev
SIZE      CREATED BY
0B        ENTRYPOINT ["/shortener"]
0B        EXPOSE [8080/tcp 9090/tcp]
0B        USER 65532:65532
16MB      COPY /out/shortener /shortener # buildkit
...                                         (the distroless base)
```

The whole image is about 24 MB, of which 16 MB is the binary. It starts in milliseconds and contains no shell:

```text
$ docker exec sh1 sh
OCI runtime exec failed: exec failed: unable to start container process: exec: "sh": executable file not found in $PATH
```

## Choosing a base image

| Base | Size | Contains | Use when |
|---|---|---|---|
| `scratch` | 0 | nothing at all | You add CA certificates and time zone data yourself (or embed them with `import _ "time/tzdata"`). |
| `gcr.io/distroless/static-debian12` | about 2 MB | CA certificates, time zone data, `/etc/passwd` with a `nonroot` user | The default for static Go binaries. |
| `alpine` | about 8 MB | a shell, `apk` | You need to debug inside the container, or the program shells out to tools. |
| `debian`/`ubuntu` slim | 30+ MB | glibc and a shell | The binary needs cgo and glibc. |

Distroless is the usual choice: everything a Go program needs (CA certificates for outgoing HTTPS, time zone data for `time.LoadLocation`) and nothing else. The "no shell" property surprises people when they first try `kubectl exec`; use `kubectl debug` with an ephemeral debug container instead, which attaches a container that has tools to the running pod.

## Run as a numeric non-root user

Containers run as root by default. If an attacker breaks out of your process, root inside the container is one step closer to root on the node. Every image should run as an unprivileged user, and Kubernetes can enforce it with `runAsNonRoot: true` (lesson 15.2).

Here is a trap this course fell into while testing the operator in lesson 15.4. The Dockerfile said `USER nonroot:nonroot`, and Kubernetes refused to start the pod:

```text
Error: container has runAsNonRoot and image has non-numeric user (nonroot), cannot verify user is non-root
```

Kubernetes cannot look inside the image's `/etc/passwd` to find out whether `nonroot` is UID 0. With a **numeric** user (`USER 65532:65532`, which is distroless's `nonroot`), it can verify. All Dockerfiles in this course now use the number.

## Multi-platform images

Your laptop may be arm64 (Apple silicon) while the cluster runs amd64, or the reverse; many clusters now run both. `docker buildx` builds one image for several platforms:

```bash
docker buildx build --platform linux/amd64,linux/arm64 -t ghcr.io/you/shortener:v1.0.0 --push .
```

For each platform, BuildKit sets `TARGETOS` and `TARGETARCH`. Because the build stage passes them to `go build`, Go **cross-compiles** each binary natively on the build machine, which takes seconds, instead of building under slow CPU emulation. This is one of Go's quiet advantages: in many other languages, a multi-platform image means emulating the other architecture for the whole build.

## ko: no Dockerfile at all

For Go services there is an even simpler option: **ko** (`github.com/ko-build/ko`) builds a Go main package straight into an image on a distroless base, without a Dockerfile and without Docker:

```bash
KO_DOCKER_REPO=ghcr.io/you ko build ./cmd/shortener
```

It uses the Go build cache, produces multi-platform and reproducible images, and is used by projects such as Knative and Tekton. The trade-off is less control over what goes into the image.

## Common mistakes

### Copying everything before downloading modules

Every code change then re-downloads every dependency. Copy `go.mod` and `go.sum` first.

### Shipping the toolchain

A single-stage build produces a 1 GB image with your source code in it. Use two stages.

### Running as root, or as a named user

Use `USER 65532:65532`, and `runAsNonRoot` in the pod.

### Forgetting CA certificates with scratch

Every outgoing HTTPS request then fails with `x509: certificate signed by unknown authority`. Use distroless, or copy the certificates.

## Key takeaways

- Go builds static binaries, so the runtime image can be almost empty: about 24 MB for the shortener, mostly the binary.
- Multi-stage Dockerfiles: build with the toolchain, copy only the binary into distroless.
- Order steps for caching and use BuildKit cache mounts; rebuilds went from 62 s to 1.6 s.
- Run as a numeric non-root user so Kubernetes can enforce `runAsNonRoot`.
- Multi-platform images build fast because Go cross-compiles; `ko` builds Go images without a Dockerfile.

## Further reading

- [Docker: multi-stage builds](https://docs.docker.com/build/building/multi-stage/)
- [Distroless images](https://github.com/GoogleContainerTools/distroless)
- [ko](https://ko.build/)
