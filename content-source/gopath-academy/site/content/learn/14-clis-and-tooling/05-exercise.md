---
title: "14.5 Exercise: a to-do CLI"
linkTitle: "14.5 Exercise: to-do CLI"
weight: 5
description: "Build a small command-line tool from a specification: global flags, subcommands with their own flags, configuration from flag or environment, file storage, and correct exit codes."
goVersion: "1.27"
---

## The task

{{< exercise dir="m14/todo" >}}
Implement `Run` in `starter.go`: a to-do list with the subcommands `add`, `list` (with `-all`) and `done`, a global `-file` flag that falls back to `$TODO_FILE` and then `todo.json`, and the exit codes 0, 1 and 2 as specified. `load` and `save` in `store.go` handle the JSON file for you.
{{< /exercise >}}

```bash
cd exercises
go test ./m14/todo
```

A session with the finished tool looks like this:

```text
$ todo add buy milk
added 1. buy milk
$ todo add write tests
added 2. write tests
$ todo done 2
$ todo list -all
1. buy milk
2. [x] write tests
$ todo done 7; echo "exit $?"
todo: no item 7 (the list has 2)
exit 1
$ todo remove 1; echo "exit $?"
todo: usage: unknown command "remove"
exit 2
```

## Hints

<details>
<summary>Hint 1: two levels of flags</summary>

Parse the global flags with one `FlagSet`; `fs.Args()` then starts with the command name. Give `list` its own `FlagSet` for `-all`. Use `flag.ContinueOnError` everywhere so `Run` decides the exit code (lesson 14.1).

</details>

<details>
<summary>Hint 2: one place for exit codes</summary>

Write an inner `run` that returns an `error`, wrap usage problems with a sentinel such as `errUsage`, and convert to an exit code once in `Run` with `errors.Is`. Do not forget `flag.ErrHelp`: `-h` is success.

</details>

<details>
<summary>Hint 3: precedence</summary>

The `-file` flag wins over `$TODO_FILE`, which wins over `todo.json`. Since the flag's default is `""`, an empty value means "not given" here, so `fs.Visit` is not needed (lesson 14.2).

</details>

The reference solution is in `exercises/m14/todo/solution.go`. Notice how `save` in `store.go` writes a temporary file and renames it: a crash in the middle of saving can never leave a half-written list, because a rename within one directory is atomic.

## Going further

- Add `todo undo N` and `todo clear` (remove done items).
- Wire it to a real `main` package and try `go install` and `go version -m` on the result (lesson 14.4).
- Rewrite it with Cobra (lesson 14.1) and compare the amount of code.
