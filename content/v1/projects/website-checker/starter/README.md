# Sitecheck starter

Implement checker.Check in checker/checker.go. The CLI, configuration, and behavior tests are provided.

```sh
go test ./...
go test -race ./...
```

Tests intentionally fail until the checker is implemented. Work through the six checkpoints on the guided project page. Do not change the expected behavior merely to make tests green.

Begin with validation and sequential requests; then add the bounded worker pool, cancellation, cleanup, and sorted results. Compare with the separate solution download only after trying the challenge.

The tests use fake HTTP transports and do not need public internet access. The actual CLI needs network access to check URLs.
