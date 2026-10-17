# miniredis

A small Redis-compatible server: the Module 8 project.

- `internal/resp`: RESP2 protocol reader and writer (with a fuzz test)
- `internal/store`: concurrency-safe keyspace with expiry and a janitor goroutine
- `internal/aof`: append-only file persistence and crash-tolerant replay
- `internal/server`: TCP server, one goroutine per client, graceful shutdown

```bash
go test -race ./...
go run ./cmd/miniredis -addr 127.0.0.1:6380 -aof /tmp/mini.aof
redis-cli -p 6380 set greeting hello
redis-cli -p 6380 get greeting
```

Supported commands: PING, ECHO, GET, SET (with EX), DEL, EXISTS, INCR, EXPIRE, TTL, DBSIZE.
