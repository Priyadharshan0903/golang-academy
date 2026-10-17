# GoPath Academy

A free, open source, in-depth Go course for developers who already know another language. From your first program to production services, concurrency, runtime internals, Kubernetes and real-world case studies.

> Working name. The final name will be chosen before the first public release.

## What makes this course different

- **Every example is tested.** Code shown on the site is embedded from real Go files in `examples/` and `projects/`, and CI runs them. If an example breaks, the build fails.
- **Exercises you run locally.** Each exercise has a starter, tests and a reference solution. Run `go test` until it passes.
- **Projects and case studies.** Build-along projects after each part, and company case studies where every claim is sourced.
- **Cloud native focus.** The site itself runs on Kubernetes, and setting that up is part of the course.

## Repository layout

```text
plan/         project plan, curriculum and progress checklists (start here)
site/         Hugo website (content, layouts, config)
examples/     tested Go code embedded in lessons
exercises/    exercises: starter code, tests and solutions (separate Go module)
projects/     build-along projects
deploy/       container image and Kubernetes setup for hosting the site
scripts/      helper scripts used by CI
```

## Quick start for learners

```bash
git clone https://github.com/Priyadharshan0903/gopath-academy.git
cd gopath-academy

go test ./examples/...                # run every example
cd exercises && go test ./m01/greeting/   # try the first exercise (fails until you solve it)
```

You need Go 1.24 or newer. The course content targets Go 1.27.

## Working on the course

Requirements: Go 1.24+, [Hugo extended](https://gohugo.io/installation/) 0.146+ (CI uses 0.168.0), and Git.

```bash
make serve    # site with live reload at http://localhost:1313
make check    # everything CI runs
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for how lessons, examples and exercises fit together, and [plan/](plan/README.md) for the roadmap and progress.

## License

- Code: [MIT](LICENSE)
- Written content: [CC BY-SA 4.0](LICENSE-CONTENT.md)
