# Lesson definition of done

Copy this list into the PR description for every new or changed lesson.

## Structure

- [ ] Front matter filled: `title`, `weight`, `description`, `goVersion`
- [ ] Starts with "What you will learn" (3 to 5 bullets)
- [ ] Explains the concept in plain words before showing code
- [ ] Says why it matters, with a real use case
- [ ] Minimal example first, deeper example second
- [ ] "Common mistakes" section with at least two real gotchas
- [ ] Exercise (if the lesson is marked `(ex)` in `curriculum.md`)
- [ ] "Key takeaways" (3 to 5 bullets)
- [ ] "Further reading" links to official docs, the spec or Go blog posts

## Correctness

- [ ] Every complete Go program or non-trivial snippet comes from `examples/` through the `example` shortcode; only short, non-compilable syntax fragments are inline
- [ ] Every example has an `Example` test with `// Output:` or a normal test, and `go test ./...` passes
- [ ] Exercise tests pass with `go test -tags solution ./...`
- [ ] Exercise starter compiles with `go vet ./...` (tests are expected to fail until solved)
- [ ] Claims about behavior checked against the spec or official docs
- [ ] Version-specific features name the Go version that introduced them

## Writing style

- [ ] Assumes the reader knows another language; does not explain what a variable is
- [ ] Compares with other languages only when it helps ("unlike Java, there are no constructors")
- [ ] Short paragraphs, active voice, second person ("you")
- [ ] No em dashes (house style)
- [ ] Shell commands in `bash` blocks, output in `text` blocks
- [ ] Reading time roughly 8 to 15 minutes

## Review

- [ ] Rendered locally with `hugo server` and read end to end
- [ ] Reviewed by Dharshan
