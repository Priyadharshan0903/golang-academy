# greet

Checkpoint project for Module 1. A small CLI that shows a properly structured Go command: `cmd/` for the binary, `internal/` for logic, flags, tests, version stamping and cross-compiled releases.

Lesson: https://github.com/Priyadharshan0903/gopath-academy/blob/main/site/content/learn/01-setup-and-tooling/08-checkpoint-project.md

```bash
make test
make build && ./bin/greet -name Ana -times 2 -shout
./bin/greet -version
make release   # binaries for linux, macOS and Windows in ./dist
```
