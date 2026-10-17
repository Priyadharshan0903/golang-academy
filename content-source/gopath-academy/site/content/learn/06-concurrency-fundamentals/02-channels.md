---
title: "6.2 Channels: unbuffered, buffered, direction"
linkTitle: "6.2 Channels"
weight: 2
description: "Send and receive values between goroutines, the difference between unbuffered and buffered channels, closing and ranging, direction types and ownership rules."
goVersion: "1.27"
---

## What you will learn

- How channels pass values **and** synchronize goroutines
- The difference between unbuffered and buffered channels
- How closing works, and how `range` and comma-ok receives detect it
- How direction types (`chan<-`, `<-chan`) enforce who sends and who receives
- The ownership rules that keep channel code correct

## The idea

A **channel** is a typed pipe between goroutines. One goroutine sends a value with `ch <- v`; another receives it with `v := <-ch`.

Channels do two jobs at once: they **move data**, and they **synchronize**, because sends and receives block until the other side is ready (or the buffer has room). Go's motto captures the design: *Do not communicate by sharing memory; instead, share memory by communicating.*

## Unbuffered channels

{{< example file="examples/m06/channels/channels_test.go" region="unbuffered" run="go test -run Example_unbuffered -v ./examples/m06/channels" >}}

`make(chan string)` creates an **unbuffered** channel. A send blocks until a receiver takes the value, and a receive blocks until a sender provides one. The two goroutines meet at that moment, like a hand-off. This makes unbuffered channels a strong synchronization tool: when the send completes, you **know** the receiver has the value.

## Buffered channels

{{< example file="examples/m06/channels/channels_test.go" region="buffered" run="go test -run Example_buffered -v ./examples/m06/channels" >}}

`make(chan int, n)` creates a channel with a buffer of `n` values. Sends block only when the buffer is full; receives block only when it is empty. `len(ch)` and `cap(ch)` report the current number of buffered values and the capacity.

Use a buffer when you know how many values will be sent and want the sender not to wait (for example, a result channel with one slot per worker), or to absorb short bursts. Do **not** add a buffer to "fix" a deadlock; that usually just moves it.

## Closing and ranging

{{< example file="examples/m06/channels/channels_test.go" region="close" run="go test -run Example_closeAndRange -v ./examples/m06/channels" >}}

- `close(ch)` signals "no more values will be sent".
- `for v := range ch` receives until the channel is closed **and** drained.
- After close, buffered values are still delivered. Then every receive returns the zero value immediately, and the comma-ok form `v, ok := <-ch` reports `ok == false`.
- Closing is optional. You only need to close when receivers need to know that the stream ended (for example, to stop a `range`).

The generator pattern in `produce` is idiomatic: a function creates a channel, starts a goroutine that sends and then closes it, and returns the channel as **receive-only**.

## Direction types

{{< example file="examples/m06/channels/channels_test.go" region="direction" run="go test -run Example_directions -v ./examples/m06/channels" >}}

- `chan<- T` can only send.
- `<-chan T` can only receive.
- A normal `chan T` converts to either automatically.

Use direction types in function signatures. They document intent and turn mistakes (receiving on an output channel, closing an input channel) into compile errors.

## Closing as a broadcast

{{< example file="examples/m06/channels/channels_test.go" region="signal" run="go test -run Example_doneSignal -v ./examples/m06/channels" >}}

A receive on a closed channel succeeds immediately, so **closing a channel wakes every goroutine waiting on it**. `chan struct{}` is the conventional type for pure signals, because the values carry no data. `context.Context`'s `Done()` channel (lesson 7.1) works exactly this way.

## The rules of channel ownership

| Operation | nil channel | closed channel |
|---|---|---|
| send `ch <- v` | blocks forever | **panics** |
| receive `<-ch` | blocks forever | returns zero value, `ok == false` |
| `close(ch)` | **panics** | **panics** |

From that table follow the rules:

1. **Only the sender closes**, and only when it is the last sender. A receiver closing a channel can make the sender panic.
2. **Close at most once.** With several senders, use a `sync.WaitGroup` and close from a separate goroutine after all senders finish (exactly what `Merge` in the exercise does).
3. **The function that creates a channel usually owns it**: it sends, closes, and hands out the receive-only end.

## Common mistakes

### Deadlock: nobody on the other side

```go
ch := make(chan int)
ch <- 1 // fatal error: all goroutines are asleep - deadlock!
```

The runtime detects when every goroutine is blocked and crashes with this message. In a larger program, a missing receiver usually shows up as a goroutine that never ends instead.

### Closing from the receiving side

Leads to `panic: send on closed channel` in a sender. Signal the sender to stop (with a separate done channel or a context) instead.

### Forgetting to close, so `range` never ends

A `for v := range ch` loop waits forever if the sender returns without closing. Use `defer close(out)` at the top of the sending goroutine.

### Using channels where a mutex is simpler

Protecting a counter or a map with a channel-based "server goroutine" is valid but often more complex than a `sync.Mutex`. Use channels to pass ownership of data or signal events; use mutexes to guard shared state (lesson 6.4).

## Exercise

{{< exercise dir="m06/pipeline" >}}
Implement `Generate`, `Square` and `Merge`. Each function creates and closes its own output channel, and every goroutine must exit when its input is drained. The last test counts goroutines before and after running 50 pipelines, so a missing `close` or a stuck goroutine will fail it. Run with `go test -race`.
{{< /exercise >}}

## Key takeaways

- Channels move values and synchronize goroutines.
- Unbuffered channels hand off directly; buffered channels decouple sender and receiver up to the buffer size.
- `close` signals the end of a stream; `range` and comma-ok receives detect it; closing wakes all receivers.
- Use direction types in signatures.
- The sender closes, exactly once; sending on or closing a closed channel panics.

## Further reading

- [Effective Go: Channels](https://go.dev/doc/effective_go#channels)
- [Share Memory By Communicating](https://go.dev/blog/codelab-share) on the Go blog
- [Channel types](https://go.dev/ref/spec#Channel_types) and [Close](https://go.dev/ref/spec#Close) in the spec
