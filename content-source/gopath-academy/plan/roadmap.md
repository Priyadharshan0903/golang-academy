# Roadmap

Two tracks run in parallel:

- **Content track**: the course itself. Quality over speed. One module at a time, reviewed before the next starts.
- **Kubernetes track**: hosting this site, used as a hands-on way to learn K8s. Details in [k8s-track.md](k8s-track.md).

Content baseline: written for **Go 1.27** (current stable as of Oct 2026). Example code must build on **Go 1.24 or newer** unless a lesson is explicitly about a newer feature, in which case the lesson says so and its example has a matching `//go:build go1.NN` constraint.

---

## Phase 0: Foundations

Content

- [x] Repo layout (`site/`, `examples/`, `exercises/`, `projects/`, `deploy/`, `plan/`)
- [x] Licenses: MIT for code, CC BY-SA 4.0 for written content
- [x] `CONTRIBUTING.md` with lesson template and style rules
- [x] Hugo site with Hextra theme (Hugo module)
- [x] `example` shortcode that embeds real, tested Go files into lessons
- [x] Examples verified with `go test` (runnable `Example` functions with `// Output:`)
- [x] Exercise pattern: starter + tests + solution behind a `solution` build tag
- [x] GitHub Actions: vet, test examples, test exercise solutions, build site
- [x] Lesson template (`site/archetypes/lesson.md`)
- [ ] Push repo to GitHub (Priyadharshan0903)
- [ ] Pick the final project name and rename
- [ ] Enable GitHub Pages as a public mirror

Kubernetes

- [x] Tiny Go static file server (`deploy/server`) with health endpoint
- [x] Multi-stage Dockerfile (Hugo build, Go server build, distroless runtime)
- [x] Server tested against the built site (health check, pages, custom 404)
- [ ] Build and run the image locally (`docker run -p 8080:8080`); Dockerfile written but not yet built

## Phase 1: Fundamentals (modules 1 to 5)

- [~] Module 1: Setup and tooling (drafted, needs review)
- [~] Module 2: Language basics (drafted, needs review)
- [~] Module 3: Composite types (drafted, needs review)
- [~] Module 4: Methods, interfaces and generics (drafted, needs review)
- [~] Module 5: Errors (drafted, needs review)

Kubernetes

- [ ] Local cluster with `kind`
- [ ] Deployment, Service, Ingress for the site
- [ ] Readiness and liveness probes using `/healthz`

## Phase 2: Concurrency and the standard library (modules 6 to 8)

- [~] Module 6: Concurrency fundamentals (drafted, needs review)
- [~] Module 7: Concurrency patterns and `context` (drafted, needs review)
- [~] Module 8: Standard library deep dives (drafted, needs review)

Kubernetes

- [ ] CI builds and pushes the image to GHCR
- [ ] Image tags from git SHA, no `latest` in manifests
- [ ] Kustomize overlays (local, prod)

## Phase 3: Testing, performance and internals (modules 9 to 11)

- [~] Module 9: Testing in depth (drafted, needs review)
- [~] Module 10: Profiling and performance (drafted, needs review)
- [~] Module 11: Runtime internals (drafted, needs review)

Kubernetes

- [ ] GitOps with Argo CD
- [ ] Resource requests and limits, HPA
- [ ] PodDisruptionBudget, rolling update strategy

## Phase 4: Production Go (modules 12 to 15)

- [~] Module 12: Building HTTP services (drafted, needs review)
- [~] Module 13: Working with data (drafted, needs review)
- [~] Module 14: CLIs and tooling (drafted, needs review)
- [~] Module 15: Go for cloud native (drafted, needs review) (Docker, K8s client-go, writing a controller)

Kubernetes

- [ ] Real cluster (k3s on a free-tier VM or similar)
- [ ] TLS with cert-manager
- [ ] Domain and DNS

## Phase 5: Projects and case studies

- [ ] Build-along projects (see [case-studies.md](case-studies.md))
- [ ] Company case studies, every claim sourced

Kubernetes

- [ ] Metrics (Prometheus) and dashboards for the site's own pods
- [ ] Structured logs, log aggregation

## Later

See [backlog.md](backlog.md).
