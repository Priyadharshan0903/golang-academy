# wordfreq

Checkpoint project for Module 3. Counts word frequencies in files or standard input and prints the top N.

```bash
go run ./cmd/wordfreq -n 5 internal/count/testdata/sample.txt
cat *.md | go run ./cmd/wordfreq -n 20
go test ./...
```
