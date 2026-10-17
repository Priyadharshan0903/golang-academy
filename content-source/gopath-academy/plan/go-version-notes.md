# Go version notes

Which Go release introduced the features this course teaches. Use it to set `goVersion` in lessons and `//go:build go1.NN` constraints in examples. Verified against the official release notes.

Course baseline: content written for **Go 1.27**; examples build on **Go 1.24+** unless gated.

| Feature | Version | Used in |
|---|---|---|
| Generics, `any`, `comparable` | 1.18 | 4.3, 4.6 |
| `errors.Join`, multiple `%w` | 1.20 | 5.4 |
| `min`, `max`, `clear` built-ins; `slices`, `maps`, `cmp` packages; `log/slog`; toolchain management (`GOTOOLCHAIN`) | 1.21 | 1.2, 3.4, 8.4 |
| Per-iteration loop variables; `range` over integers; enhanced `ServeMux` patterns; `slices.Delete` zeroes tail | 1.22 | 2.5, 2.6, 3.3, 12.1 |
| Range over functions, `iter` package, `maps.Keys` returning iterators | 1.23 | 3.8 |
| `tool` directive in `go.mod`; `strings.Lines`, `SplitSeq`, `FieldsSeq`; `testing.B.Loop`; `T.Context`; `os.Root`; Swiss table maps; `go build` stamps VCS version; `crypto/rand.Read` never returns an error | 1.24 | 1.4, 1.6, 3.8, 9.x, 11.6 |
| `sync.WaitGroup.Go`; `testing/synctest` (GA); container-aware `GOMAXPROCS`; trace `FlightRecorder`; `net/http.CrossOriginProtection`; `encoding/json/v2` (experiment, `GOEXPERIMENT=jsonv2`); vet `waitgroup` and `hostport` analyzers | 1.25 | 6.4, 9.x, 10.3, 11.1, 12.6, 8.2 |
| `errors.AsType`; `new(expr)`; Green Tea GC on by default; rewritten `go fix` with modernizers; `log/slog.NewMultiHandler`; goroutine leak profile (experiment); `reflect` iterators; `bytes.Buffer.Peek`; `go mod init` writes previous minor version | 1.26 | 1.7, 5.2, 7.6, 8.4, 11.3 |
| Generic methods; goroutine leak profile GA (`/debug/pprof/goroutineleak`); `encoding/json/v2` and `encoding/json/jsontext` GA, with `encoding/json` backed by v2; new `uuid` package; `strings.CutLast`, `bytes.CutLast`; `testing/synctest.Sleep`; `httptest.NewTestServer` (in-memory network); `crypto/mldsa`; `net/url.URL.Clone`; `go test` runs `stdversion` vet check; time channels always synchronous (`asynctimerchan` removed); size-specialized small allocations | 1.27 | 4.7, 6.3, 7.6, 8.2, 8.6, 9.x |

## Testing newer features

- Examples that use a feature newer than 1.24 go in their own file with a matching `//go:build go1.NN` line.
- `gofmt` from an older release cannot parse newer syntax, so CI runs the style check on stable Go only.
- Packages where every file is gated are skipped on older toolchains when tests are run with `./...`.

## Sources

- https://go.dev/doc/go1.25
- https://go.dev/doc/go1.26
- https://go.dev/doc/go1.27
- https://go.dev/doc/devel/release

## Verified for Module 10

- Go 1.25 and Go 1.26 both: "The compiler can now allocate the backing store for slices on the stack in more situations" (release notes).
- Go 1.26: the `pprof` web UI (`-http`) defaults to the flame graph view; the graph view is under View, Graph.
- Go 1.25: `trace.FlightRecorder`, configured with `trace.FlightRecorderConfig{MinAge, MaxBytes}`.

## Verified for Module 11

- Go 1.25 container-aware GOMAXPROCS: on Linux the default respects the cgroup CPU bandwidth limit (Kubernetes "CPU limit", not "CPU requests"), and the runtime periodically updates GOMAXPROCS. Disabled by setting GOMAXPROCS manually, or with GODEBUG `containermaxprocs=0` / `updatemaxprocs=0`. New `runtime.SetDefaultGOMAXPROCS`.
- Go 1.25 Green Tea GC experiment (`GOEXPERIMENT=greenteagc`); Go 1.26 default, 10 to 40% less GC overhead expected for GC-heavy programs, extra ~10% with vector instructions on newer amd64. Opt-out `GOEXPERIMENT=nogreenteagc`, "expected to be removed in Go 1.27"; checked: the go1.27.2 toolchain still accepts it.
- Go 1.24 Swiss table maps; `GOEXPERIMENT=noswissmap` accepted by go1.24.7, rejected ("unknown GOEXPERIMENT") by go1.27.2.
- Go 1.26 runtime/metrics: `/sched/goroutines` state counts, `/sched/threads:threads`, `/sched/goroutines-created:goroutines`.
- Go 1.27: size-specialized malloc for small (<80 byte) allocations, opt-out `GOEXPERIMENT=nosizespecializedmalloc`.
