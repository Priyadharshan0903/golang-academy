# Examples

Every code sample on the website is embedded from a file in this folder, and every file here is tested.

```bash
go test ./examples/...          # all examples
go test -v ./examples/m01/...   # module 1, verbose
go run ./examples/m01/hello     # run a single program
```

Folders are named `mNN/<topic>`, where `NN` is the module number.
