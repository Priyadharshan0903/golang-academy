# Contributing

Thanks for helping make this course better. This guide explains how the pieces fit together.

## How a lesson is built

A lesson is made of three parts that live in different folders:

| Part | Location | Checked by |
|---|---|---|
| Lesson text | `site/content/learn/<module>/<lesson>.md` | Review, `hugo` build |
| Example code | `examples/mNN/<topic>/` | `go vet`, `go test` |
| Exercise | `exercises/mNN/<topic>/` | `go vet`, `go test -tags solution` |

### Embedding examples

Never paste a complete program into Markdown. Put it in `examples/` with a test, then embed it:

```text
{{< example file="examples/m01/greeting/greeting.go" >}}
{{< example file="examples/m01/greeting/greeting.go" region="visibility" run="go test ./examples/m01/greeting" >}}
```

To embed part of a file, mark a region in the Go source. Leave a blank line after the start marker so it never becomes part of a doc comment:

```go
// [start visibility]

func Hello(name string) string { ... }

// [end visibility]
```

Marker lines are never shown on the site. The build fails if a file or region does not exist.

Short fragments that only illustrate syntax and cannot compile on their own (a few lines) may be written inline in Markdown.

### Testing examples

Prefer testable examples, because they double as documentation:

```go
func ExampleHello() {
	fmt.Println(greeting.Hello("Ana"))
	// Output: Hello, Ana!
}
```

Use normal tests when the output depends on the machine (time, OS, randomness).

### Writing exercises

Each exercise folder contains:

- `starter.go` with `//go:build !solution`: what learners edit. It must compile.
- `solution.go` with `//go:build solution`: the reference solution.
- `*_test.go`: tests shared by both.

Embed it in a lesson with:

```text
{{< exercise dir="m01/greeting" >}}
What to implement.
{{< /exercise >}}
```

Name the exercise folder after its package (`m01/tempconv` contains `package tempconv`).

## Style

- Write for developers who know another language. Do not explain what a loop is.
- Short paragraphs, second person, active voice.
- **No em dashes** anywhere. `scripts/check-style.sh` fails CI if it finds one.
- Every lesson follows `site/archetypes/lesson.md` and passes [plan/lesson-checklist.md](plan/lesson-checklist.md).
- Claims about behavior must match the spec, official docs or release notes. Link them in "Further reading".

Create a new lesson from the template:

```bash
cd site
hugo new content --kind lesson learn/02-language-basics/01-variables.md
```

## Before opening a pull request

```bash
make check
```

This runs the style checks, `go vet`, all tests (including exercise solutions) and a full site build.
