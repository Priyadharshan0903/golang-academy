---
title: "8.8 Project: a mini Redis server"
linkTitle: "8.8 Project: mini Redis"
weight: 8
description: "Build a Redis-compatible TCP server with only the standard library: the RESP protocol with bufio, a concurrent keyspace with expiry, append-only file persistence, and graceful shutdown. Works with redis-cli."
goVersion: "1.27"
---

## What you will build

A small server that speaks the real Redis protocol, so the real `redis-cli` can talk to it:

```bash
go run ./projects/miniredis/cmd/miniredis -addr 127.0.0.1:6380 -aof /tmp/mini.aof
```

```bash
$ redis-cli -p 6380 set greeting "hello gopher"
OK
$ redis-cli -p 6380 get greeting
"hello gopher"
$ redis-cli -p 6380 incr visits
(integer) 1
```

Stop it with Ctrl+C, start it again, and the data is still there.

It brings together Modules 6, 7 and 8: TCP networking, `bufio` parsing, a mutex-protected store, background goroutines, `context` cancellation, files and `slog`. The code is in [`projects/miniredis`](https://github.com/Priyadharshan0903/gopath-academy/tree/main/projects/miniredis).

## Architecture

```text
              ┌───────────────── server ──────────────────┐
redis-cli ──▶ │ accept loop ──▶ one goroutine per client  │
redis-cli ──▶ │                   │  resp.Reader / Writer  │
              │                   ▼                        │
              │              execute(cmd) ──▶ store (RWMutex, TTLs)
              │                   │                        │
              │                   └──▶ aof.Log (append writes)
              │ janitor goroutine: removes expired keys    │
              └────────────────────────────────────────────┘
```

Four packages, each with one job:

| Package | Responsibility |
|---|---|
| `internal/resp` | Parse commands and write replies in RESP2 |
| `internal/store` | The keyspace: values, expiry, concurrency safety |
| `internal/aof` | Append write commands to a file; replay it on startup |
| `internal/server` | Accept connections, dispatch commands, shut down gracefully |

## The protocol: `internal/resp`

Redis clients send each command as an **array of bulk strings**, each prefixed with its length:

```text
*3\r\n$3\r\nSET\r\n$8\r\ngreeting\r\n$12\r\nhello gopher\r\n
```

Length prefixes make the protocol **binary safe**: a value can contain spaces, newlines or zero bytes.

{{< example file="projects/miniredis/internal/resp/resp.go" >}}

Things to notice:

- `bufio.Reader.ReadSlice('\n')` reads a line without allocating, and `io.ReadFull` reads exactly `size+2` bytes for a bulk string.
- **Limits** on array length, bulk size and line length stop a malicious client from making the server allocate gigabytes.
- **Errors are classified**: `ErrProtocol` for bad input (the server replies with an error and closes), `io.EOF` for a clean disconnect between commands, and `io.ErrUnexpectedEOF` for a disconnect in the middle of one.
- The writer buffers replies, and the server calls `Flush` once per command.

The tests include a **fuzz test** (Module 9 explains fuzzing): `go test -fuzz=FuzzReadCommand` throws millions of random inputs at the parser to make sure it never panics or hangs.

## The keyspace: `internal/store`

{{< example file="projects/miniredis/internal/store/store.go" >}}

- **One `sync.RWMutex`** guards the map. `GET`, `EXISTS` and `TTL` take the read lock, so many clients can read in parallel.
- **`INCR` is one critical section.** Reading, parsing, adding and writing back under one lock is what makes 1000 concurrent increments add up to exactly 1000 (the server test checks this). Doing it as `GET` then `SET` would lose updates.
- **Lazy expiry plus a janitor.** Reads treat expired keys as missing without deleting them (they only hold the read lock). `RunJanitor` periodically removes expired keys so memory is reclaimed even for keys nobody reads again. Redis uses the same two-part approach.
- **Injected clock** for deterministic expiry tests.

## Persistence: `internal/aof`

{{< example file="projects/miniredis/internal/aof/aof.go" >}}

An **append-only file** records every write command in RESP format. On startup, replaying the file rebuilds the data. Design decisions:

- **Only successful writes are appended** (a `DEL` that deleted nothing is skipped).
- **fsync policy.** Calling `Sync` after every write is safest and slowest. The default here, like Redis' default, syncs once per second, so a crash can lose up to a second of writes.
- **Truncated tail tolerance.** If the process crashed in the middle of an append, the last command is incomplete. Replay stops cleanly at that point instead of refusing to start.

## The server: `internal/server`

{{< example file="projects/miniredis/internal/server/server.go" >}}

- **One goroutine per connection** is the idiomatic Go server design. Goroutines are cheap, and blocking reads are simple to write; the runtime multiplexes them over a few threads with an efficient network poller underneath.
- **Graceful shutdown.** `context.AfterFunc` closes the listener when the context is cancelled, which makes `Accept` return. The server then closes every tracked client connection (unblocking their reads) and waits for their goroutines with a `WaitGroup`. The test `TestShutdownClosesClients` checks that idle clients cannot block shutdown.
- **Panic isolation.** A deferred `recover` in each connection handler means a bug triggered by one client's command does not crash the server for everyone (lesson 5.5).
- **Errors as replies.** Wrong argument counts and unknown commands get Redis-style `-ERR` replies, so real clients show sensible messages.

## The main program

{{< example file="projects/miniredis/cmd/miniredis/main.go" >}}

On startup it replays the AOF into the store **through a server without an AOF attached**, so replayed commands are not appended to the file again. Then it starts the janitor and the once-per-second sync goroutine, and serves until SIGINT or SIGTERM.

## Testing

```bash
go test -race ./projects/miniredis/...
go test -fuzz=FuzzReadCommand -fuzztime=30s ./projects/miniredis/internal/resp
```

The server tests start a real server on a random local port and talk to it with a tiny RESP client, including 20 concurrent clients incrementing one counter, inline commands like you would type in `telnet`, and shutdown with idle connections. The `main` test simulates a restart by writing through one server and replaying into a fresh store.

## Stretch goals

1. **Absolute expiry in the AOF.** `SET k v EX 100` is replayed with a fresh 100 seconds. Rewrite it in the log as `SET k v` plus `PEXPIREAT k <unix-ms>` so restarts do not extend TTLs.
2. **AOF rewriting.** The log grows forever. Add a `BGREWRITEAOF` command that writes the current dataset to a new file and swaps it in atomically (lesson 8.5).
3. **More commands.** `MGET`, `MSET`, `KEYS pattern` (use `path.Match`), `APPEND`, `GETDEL`.
4. **Sharded locks.** Split the store into 16 maps, each with its own mutex, chosen by a hash of the key. Benchmark it against the single lock in Module 10.
5. **Pipelining.** Real clients often send many commands before reading replies. The server already handles it; measure the speedup with `redis-benchmark -p 6380 -P 16`.

## Module 8 recap

You can now:

- Build streaming programs from `io` and `bufio` pieces
- Encode and decode JSON correctly with `encoding/json`, and use `encoding/json/v2` on Go 1.27
- Handle time, durations and time zones without the classic bugs
- Log with structured `slog` records
- Work with files safely, embed assets and confine paths with `os.Root`
- Format and convert text efficiently
- Make HTTP requests that cannot hang or leak

Next is **Part 3: Quality, performance and internals**, starting with **[Module 9: Testing in depth](/learn/09-testing/)**.
