---
title: "2.7 defer, and how it is evaluated"
linkTitle: "2.7 defer"
weight: 7
description: "How defer schedules cleanup, the exact moment its arguments are evaluated, how it interacts with named results and panics, and the traps of defer in loops."
goVersion: "1.27"
---

## What you will learn

- What `defer` does and the order deferred calls run in
- Exactly **when** the arguments of a deferred call are evaluated
- The standard acquire, check, defer release pattern
- How deferred functions can change named results
- The traps: `defer` in loops, ignored `Close` errors and `os.Exit`

## The idea

`defer` schedules a function call to run **when the surrounding function returns**, however it returns: a normal `return`, falling off the end, or a panic. It is Go's answer to `try / finally`, `using` blocks and destructors.

The big benefit is locality. You write the cleanup right next to the thing being cleaned up, instead of at every return statement.

## Order: last in, first out

Deferred calls are pushed onto a stack. They run in **reverse** order:

{{< example file="examples/m02/deferring/deferring_test.go" region="order" run="go test -run Example_order -v ./examples/m02/deferring" >}}

Reverse order is exactly what you want for cleanup: resources acquired later are released first, like closing a file before releasing the lock that protected it.

## When arguments are evaluated

This is the rule that trips people up. A deferred call's **function value and arguments are evaluated immediately**, at the `defer` statement. Only the call itself is delayed:

{{< example file="examples/m02/deferring/deferring_test.go" region="args" run="go test -run Example_argumentsEvaluatedEarly -v ./examples/m02/deferring" >}}

If you want the deferred code to see the **latest** value, defer a closure. The closure captures the variable and reads it when it runs:

{{< example file="examples/m02/deferring/deferring_test.go" region="closure" run="go test -run Example_closureSeesLatest -v ./examples/m02/deferring" >}}

## The standard cleanup pattern

Acquire the resource, check the error, then defer the release **immediately**:

{{< example file="examples/m02/deferring/deferring_test.go" region="cleanup" run="go test -run Example_cleanup -v ./examples/m02/deferring" >}}

The order matters. If you write `defer f.Close()` **before** checking `err`, then on failure `f` is `nil` and the deferred `Close` runs on a nil file.

You will use the same shape everywhere:

```go
mu.Lock()
defer mu.Unlock()

resp, err := http.Get(url)
if err != nil {
	return err
}
defer resp.Body.Close()
```

## Changing results from a defer

A deferred **closure** can read and modify the function's **named** results, because they are ordinary variables that are still in scope. The deferred code runs after the `return` statement has set them, but before the caller sees them:

{{< example file="examples/m02/deferring/deferring_test.go" region="named" run="go test -run Example_modifyNamedResult -v ./examples/m02/deferring" >}}

The example uses `recover` to turn a panic into an error. Panics and `recover` are covered properly in Module 5; the point here is the mechanism: **deferred functions run even when the function panics**, and they can adjust named results.

A gentler, very common use is adding context to any error a function returns:

```go
func process(path string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("process %s: %w", path, err)
		}
	}()
	// ... many return statements, all wrapped automatically
}
```

## Cost

`defer` used to have a noticeable overhead, which led to advice to avoid it in hot paths. Since Go 1.14, most defers are compiled inline ("open-coded defers") and cost close to nothing. Use `defer` for correctness by default, and only remove one if a profile (Module 10) shows it matters.

## Common mistakes

### `defer` inside a loop

```go
for _, path := range paths {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close() // nothing closes until the WHOLE function returns
	// ...
}
```

With thousands of files, you run out of file descriptors. Move the loop body into its own function, so each `defer` runs at the end of each iteration:

```go
for _, path := range paths {
	if err := processFile(path); err != nil { // processFile defers its own Close
		return err
	}
}
```

### Ignoring the error from `Close` on a file you wrote

For files opened for **reading**, ignoring `Close` errors is fine. For files you **wrote**, `Close` (or `Sync`) can be the moment you learn the data did not reach the disk. Return that error, for example with the named result pattern above:

```go
defer func() {
	if cerr := f.Close(); cerr != nil && err == nil {
		err = cerr
	}
}()
```

### Expecting defers to run on `os.Exit`

`os.Exit` ends the program immediately. **Deferred calls do not run**, and neither does any other cleanup. The same applies to `log.Fatal`, which calls `os.Exit`. This is one reason the checkpoint project in Module 1 kept all logic in `run()` and called `os.Exit` only in `main`.

### Deferring before the error check

`defer f.Close()` before `if err != nil` defers a call on a nil value. Always check first.

## Key takeaways

- `defer` runs a call when the surrounding function returns, including on panic.
- Deferred calls run in last in, first out order.
- Arguments are evaluated at the `defer` statement; use a closure to see later values.
- Pattern: acquire, check the error, `defer` the release on the next line.
- Deferred closures can modify named results, which is useful for wrapping errors.
- Avoid `defer` in loops, handle `Close` errors for written files, and remember that `os.Exit` skips defers.

## Further reading

- [Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover) on the Go blog
- [Defer statements](https://go.dev/ref/spec#Defer_statements) in the spec
- [Effective Go: Defer](https://go.dev/doc/effective_go#defer)
