# Decisions

Short records of the choices we made. Add a new entry instead of editing an old one when we change our minds.

---

## D1. Static site with Hugo

Date: 2026-10-16. Status: accepted.

- **Chose**: Hugo.
- **Alternatives**: Astro Starlight, Next.js, Docusaurus, a custom Go server rendering Markdown.
- **Why**: Single binary, very fast builds, content is plain Markdown so contributions are simple PRs. Hugo is written in Go, which fits the project. No JavaScript build chain to maintain.
- **Trade-off**: Hugo templates are less flexible than a JS framework when we add login and interactivity later. We accept that; those features can be a separate service.

## D2. Hextra theme as a Hugo module

Date: 2026-10-16. Status: accepted.

- **Chose**: Hextra, imported as a Hugo module (pinned in `site/go.mod`).
- **Why**: Docs-style sidebar, built-in search, dark mode, callouts, tabs, steps. MIT licensed.
- **Trade-off**: Building the site requires Go installed (Hugo modules use the Go toolchain). Every contributor has Go anyway.

## D3. Examples are real Go files, tested in CI

Date: 2026-10-16. Status: accepted.

- **Chose**: All non-trivial code in lessons lives in `examples/` as Go packages with `Example` functions (`// Output:` comments) or tests. Lessons embed them with the `example` shortcode.
- **Why**: Code in Markdown rots silently. Code in tested files fails CI the moment it breaks.
- **Trade-off**: Writing a lesson takes slightly longer. Worth it.

## D4. Exercises with a `solution` build tag

Date: 2026-10-16. Status: accepted.

- **Chose**: Each exercise has `starter.go` (`//go:build !solution`), `solution.go` (`//go:build solution`) and a shared `_test.go`.
- **Why**: Learners run `go test` and see failures until they solve it. CI runs `go test -tags solution` to prove the tests and the reference solution agree, and `go vet` proves the starter compiles.
- **Trade-off**: Solutions are visible in the repo. This is an open course, so that is fine; the exercise README asks learners not to peek.

## D5. Licenses

Date: 2026-10-16. Status: accepted.

- Code (`examples/`, `exercises/`, `projects/`, `deploy/`): MIT.
- Written content (`site/content/`): CC BY-SA 4.0.

## D6. Hosting on Kubernetes, with GitHub Pages as a mirror

Date: 2026-10-16. Status: accepted.

- **Chose**: Serve the built site from a small Go file server in a container on Kubernetes.
- **Why**: Hosting the course is also the K8s learning track. GitHub Pages stays as a free, always-on mirror in case the cluster is down or costs become an issue.

## D7. Go version policy

Date: 2026-10-16. Status: accepted.

- Content is written against the current stable release (Go 1.27 at the time of writing).
- Example code must build on Go 1.24 or newer, unless a lesson is specifically about a newer feature. Those examples carry a `//go:build go1.NN` constraint.
- Review the course after each Go release (February and August) and update `goVersion` in lesson front matter.

## D8. Try it locally, no in-browser runner for v1

Date: 2026-10-16. Status: accepted.

- Learners clone the repo and run examples and exercises locally.
- An in-browser runner is in [backlog.md](backlog.md).

## D9. House style: no em dashes

Date: 2026-10-16. Status: accepted.

- No em dash characters anywhere in content, code or docs. CI fails if one appears.

## D10. Vendor the search library

Date: 2026-10-16. Status: accepted.

- **Chose**: Keep a copy of FlexSearch (Apache-2.0) in `site/assets/lib/flexsearch/` instead of letting the theme fetch it from a CDN at build time.
- **Why**: Builds are reproducible and work offline or behind restrictive networks, including CI and container builds.
- **Trade-off**: We update it by hand when the theme moves to a new FlexSearch version.

## D11. Projects with dependencies are separate modules

**Decision:** The main module stays dependency-free. Code that needs third party packages lives in its own module: `examples/thirdparty` (examples), `projects/shortener` (from milestone 2: Prometheus client, `golang.org/x/time/rate`; from milestone 3: pgx and go-redis) `projects/jobqueue` (pgx) and `projects/shortlink-operator` (controller-runtime, client-go, with controller-gen as a `go tool`). Database tests skip unless `TEST_DATABASE_URL` / `TEST_REDIS_ADDR` are set; CI provides Postgres 16 and Redis 7 as service containers. Each has its own CI job that runs `go mod tidy`, `go vet` and `go test -race`, and no committed `go.sum`.

**Why:** Readers can run almost every example with nothing but a Go toolchain, while the projects still use the libraries real services use. The shortener needs a database driver by Module 13, so it moved to its own module in Module 12.

**Trade-off:** These modules require a newer Go (`go 1.26.0`, set by the Prometheus client) and are tested on stable Go only.
