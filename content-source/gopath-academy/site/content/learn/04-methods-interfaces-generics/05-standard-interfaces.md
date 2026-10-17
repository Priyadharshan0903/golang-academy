---
title: "4.5 Common standard interfaces"
linkTitle: "4.5 Standard interfaces"
weight: 5
description: "fmt.Stringer, io.Reader, io.Writer and sort.Interface: the small interfaces that make Go's standard library composable, and how to implement them yourself."
goVersion: "1.27"
---

## What you will learn

- How `fmt.Stringer` controls how your types print
- The `io.Writer` and `io.Reader` contracts, and how to implement them correctly
- Why these two tiny interfaces make streams composable
- How `sort.Interface` works, and why `slices.SortFunc` usually replaces it
- Other interfaces you will meet soon

## The idea

A handful of one-method interfaces in the standard library are implemented by hundreds of types. Learn them, and your own types instantly work with `fmt`, `io.Copy`, compression, hashing, HTTP bodies, files and more.

## `fmt.Stringer`

```go
type Stringer interface {
	String() string
}
```

When `fmt` prints a value with `%v`, `%s` or `Println`, it checks whether the value is a `Stringer` and calls `String()`:

{{< example file="examples/m04/stdinterfaces/stdinterfaces_test.go" region="stringer" run="go test -run Example_stringer -v ./examples/m04/stdinterfaces" >}}

This finishes the `Weekday` enum from lesson 2.3. For long enums, generate `String()` with the `stringer` tool (lesson 1.4) instead of writing the switch by hand.

{{< callout type="warning" >}}
Never call `fmt.Sprintf("%v", x)` on the receiver itself inside `String()`. It calls `String()` again and recurses until the stack overflows. Convert to the underlying type first, as `int(d)` does above.
{{< /callout >}}

## `io.Writer`

```go
type Writer interface {
	Write(p []byte) (n int, err error)
}
```

The contract: write `len(p)` bytes from `p` and return how many were written. If `n < len(p)`, `err` must be non-nil. `Write` must not modify `p` or keep it after returning.

Files, network connections, `bytes.Buffer`, `strings.Builder`, HTTP response writers, hashers and gzip compressors are all writers. A **wrapper** that adds behavior to any writer is a few lines:

{{< example file="examples/m04/stdinterfaces/stdinterfaces_test.go" region="writer" run="go test -run Example_writer -v ./examples/m04/stdinterfaces" >}}

`fmt.Fprintf` and `io.WriteString` accept any `io.Writer`, so they work with `CountingWriter` without knowing it exists.

## `io.Reader`

```go
type Reader interface {
	Read(p []byte) (n int, err error)
}
```

The contract is subtle, and worth reading twice:

- `Read` fills **up to** `len(p)` bytes and returns how many it read. Fewer bytes than requested is normal.
- At the end of the data it returns `io.EOF`. It may return `n > 0` **together with** `io.EOF` on the last call.
- So callers must **process the `n` bytes first, then look at the error.**

{{< example file="examples/m04/stdinterfaces/stdinterfaces_test.go" region="reader" run="go test -run Example_reader -v ./examples/m04/stdinterfaces" >}}

`upperReader` processes `p[:n]` and passes the error through, which keeps it correct for both cases. In practice you rarely call `Read` yourself; you hand readers to `io.Copy`, `io.ReadAll`, `bufio.Scanner` or `json.NewDecoder`, which handle the loop. Module 8 goes much deeper.

### Why this matters

Because readers and writers are so small, they **stack**. A typical pipeline in real code:

```go
f, _ := os.Open("access.log.gz")        // io.Reader: file
gz, _ := gzip.NewReader(f)              // io.Reader: decompressing
sc := bufio.NewScanner(gz)              // reads lines from it
```

Each layer only knows that the layer below is an `io.Reader`. Swap the file for an HTTP response body and nothing else changes.

## `sort.Interface`

```go
type Interface interface {
	Len() int
	Less(i, j int) bool
	Swap(i, j int)
}
```

{{< example file="examples/m04/stdinterfaces/stdinterfaces_test.go" region="sort" run="go test -run Example_sortInterface -v ./examples/m04/stdinterfaces" >}}

This is the pre-generics way to sort, and you will see it in older code and in data structures like `container/heap` (which builds on it). For sorting a slice today, `slices.SortFunc` (lesson 3.4) is shorter, faster and type safe.

## Interfaces you will meet next

| Interface | Package | Lesson |
|---|---|---|
| `error` | built in | Module 5 |
| `io.Closer`, `io.ReadCloser`, `io.WriterTo`, `io.ReaderFrom` | `io` | 8.1 |
| `json.Marshaler`, `json.Unmarshaler` | `encoding/json` | 8.2 |
| `encoding.TextMarshaler` | `encoding` | 8.2 |
| `http.Handler` | `net/http` | 12.1 |
| `context.Context` | `context` | 7.1 |
| `heap.Interface` | `container/heap` | (builds on `sort.Interface`) |

## Common mistakes

### Recursion in `String()`

Using `%v` or `%s` on the receiver inside its own `String` method. Convert to the underlying type.

### Ignoring `n` when `Read` returns an error

Dropping the last chunk of data because `err == io.EOF` arrived with it. Handle `p[:n]` first.

### Treating `io.EOF` as a failure

`io.EOF` means "done", not "broken". Helpers like `io.ReadAll` and `bufio.Scanner` hide it from you; code that loops over `Read` itself must check `errors.Is(err, io.EOF)`.

### Defining a `String()` with a pointer receiver on a value type

`fmt` receives a copy, which does not have pointer methods, so your `String` is ignored. Use a value receiver for `String()` on small types.

## Key takeaways

- `fmt.Stringer` controls printing; avoid recursion inside `String()`.
- `io.Writer` and `io.Reader` are the backbone of streaming in Go, and wrappers compose freely.
- `Read` may return data and `io.EOF` together: process `n` bytes before checking the error.
- `sort.Interface` still exists, but `slices.SortFunc` is the modern way to sort slices.

## Further reading

- [Package io](https://pkg.go.dev/io), especially the Reader and Writer docs
- [Package fmt: Stringer](https://pkg.go.dev/fmt#Stringer)
- [Package sort](https://pkg.go.dev/sort)
