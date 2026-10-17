# Exercises

Each folder is one exercise. Edit `starter.go` until the tests pass.

```bash
cd exercises
go test ./m01/greeting/      # one exercise
go test ./m01/...            # all exercises in module 1
```

The tests fail until your code is correct. That is expected.

Every exercise also has a `solution.go`, which is only compiled with the `solution` build tag. Try the exercise yourself before reading it. To check the reference solution:

```bash
go test -tags solution ./m01/greeting/
```

| Module | Exercise | Lesson |
|---|---|---|
| 1 | `m01/greeting` | 1.3 Your first program |
| 1 | `m01/tempconv` | 1.5 Packages and layout |
| 1 | `m01/cli` | 1.8 Checkpoint project |
| 2 | `m02/palindrome` | 2.4 Strings, bytes and runes |
| 2 | `m02/stats` | 2.6 Functions and closures |
| 2 | `m02/ptr` | 2.8 Pointers |
| 3 | `m03/ints` | 3.2 Slices |
| 3 | `m03/inventory` | 3.5 Maps |
| 3 | `m03/seq` | 3.8 Iterators |
| 4 | `m04/kv` | 4.2 Interfaces |
| 4 | `m04/set` | 4.6 Generics |
| 5 | `m05/config` | 5.2 Wrapping errors |
| 5 | `m05/form` | 5.4 Joining errors |
| 6 | `m06/pipeline` | 6.2 Channels |
| 6 | `m06/parallel` | 6.4 The sync package |
| 7 | `m07/poll` | 7.1 context |
| 7 | `m07/pmap` | 7.2 Worker pools |
| 8 | `m08/grep` | 8.1 io and bufio |
| 8 | `m08/apiclient` | 8.7 HTTP client |
| 9 | `m09/wordwrap` | 9.1 Table-driven tests (write the tests) |
| 9 | `m09/fizzhttp` | 9.5 httptest |
| 9 | `m09/rle` | 9.6 Fuzzing |
| 10 | `m10/hashtags` | 10.7 Optimize under a budget |
| 11 | `m11/layout` | 11.7 Shrink a struct |
| 12 | `m12/notes` | 12.7 A notes JSON API |
| 13 | `m13/readthrough` | 13.6 A read-through cache |
| 14 | `m14/todo` | 14.5 A to-do CLI |
| 15 | `m15/drain` | 15.6 Drain gracefully |
