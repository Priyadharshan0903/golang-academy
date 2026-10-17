# Backlog

Agreed ideas that are deliberately postponed. Move an item into [roadmap.md](roadmap.md) when we decide to schedule it.

## Learner features

- [ ] Login (GitHub OAuth is the natural fit for developers)
- [ ] Progress tracking per lesson and module
- [ ] In-browser code runner (options: embed the Go Playground, a self-hosted sandbox, or an interpreter compiled to WebAssembly)
- [ ] Quizzes at the end of each module
- [ ] Certificates of completion (maybe)

## Content

- [ ] Cheat sheets per module (printable)
- [ ] "Go for Java / Python / JavaScript developers" comparison pages
- [ ] Glossary
- [ ] Interview preparation section
- [ ] Translations (Tamil, Hindi) once the English content is stable

## Site

- [ ] Custom landing page design
- [ ] Open Graph images per lesson
- [ ] Analytics that respect privacy (e.g. a self-hosted option)
- [ ] Broken link checker in CI

## Build

- [ ] Commit `go.sum` for `examples/thirdparty`, `projects/shortener`, `projects/jobqueue` and `projects/shortlink-operator`: run `go mod tidy` on a machine with access to the Go module proxy, commit the files, and remove them from `.gitignore` (they were not generated in the authoring sandbox, which had no proxy access; see D11)
- [ ] Add Dependabot or Renovate for the Go modules and GitHub Actions
