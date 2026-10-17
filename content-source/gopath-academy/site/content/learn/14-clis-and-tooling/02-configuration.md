---
title: "14.2 Configuration for CLIs: files, environment and flags"
linkTitle: "14.2 Configuration for CLIs"
weight: 2
description: "Give a CLI layered configuration with a predictable order of precedence: defaults, a config file in the user's config directory, environment variables and flags, using flag.Visit to know which flags were actually set."
goVersion: "1.27"
---

## What you will learn

- The standard order of precedence for CLI configuration
- Where config files belong: `os.UserConfigDir`
- Telling "flag not given" apart from "flag given with its default value" with `flag.Visit`
- Strict config files that reject typos
- Showing users where each setting came from

## The idea

A service gets its configuration from the environment (lesson 12.4). A CLI is used by people, interactively and in scripts, so it usually accepts several sources, each overriding the previous one:

```text
defaults  <  config file  <  environment variables  <  command-line flags
```

The logic: the config file holds a user's long-term preferences, environment variables adjust them for a shell session or a CI job, and a flag is the most specific, deliberate choice for one run. `git`, `kubectl`, `aws` and most well-behaved tools follow this order.

## The loader

{{< example file="examples/m14/layered/layered.go" region="config" >}}

The interesting line is **`fs.Visit`**. A flag's default value cannot be told apart from a value the user typed by looking at the variable: if `-timeout` defaults to `0` and the environment says `5s`, did the user type `-timeout 0`? `fs.Visit` calls its function only for flags that were **set on the command line**, while `fs.VisitAll` visits all of them. Applying flags inside `Visit` means an unset flag never overrides the file or the environment.

That is also why the flags are declared with empty defaults, and the real defaults live in the `Config` literal: the help text explains them in words instead.

## The config file

{{< example file="examples/m14/layered/layered.go" region="file" >}}

- **Location.** `os.UserConfigDir()` returns the platform's convention: `$XDG_CONFIG_HOME` or `~/.config` on Linux, `~/Library/Application Support` on macOS, `%AppData%` on Windows. Put files in a subdirectory named after the tool.
- **Optional, unless asked for.** A missing default file is normal. A missing file the user named explicitly (with `-config` or `SHORTCTL_CONFIG`) is an error.
- **Pointers for "not set".** `*string` fields distinguish a key that is absent from a key set to `""`.
- **Strict.** `DisallowUnknownFields` turns `"sever": "..."` into an error instead of a silently ignored setting.

JSON keeps this dependency-free. Many tools prefer YAML or TOML for config files people edit by hand, which need a third party package; the layering logic is the same.

## Precedence, demonstrated

{{< example file="examples/m14/layered/layered_test.go" region="example" run="go test -v -run ExampleLoad ./examples/m14/layered" >}}

The file sets the server and a 3s timeout, the environment overrides the timeout to 5s, and a flag sets the output format. Recording the **source** of each value costs a few lines and saves a lot of confused debugging: a `shortctl config` command that prints this table is a feature users love ("why is it talking to staging?").

## Common mistakes

### Flag defaults overriding the environment

Applying every flag value, set or not, silently discards the file and environment. Use `fs.Visit`.

### Config files in the home directory root

`~/.mytoolrc` clutters home directories. Use `os.UserConfigDir()`.

### Ignoring unknown keys

Typos then do nothing, silently. Reject them.

### Secrets in flags

Command lines are visible to other users of the machine (`ps`) and saved in shell history. Read tokens from environment variables, files, or the OS keychain.

## Key takeaways

- Order sources as defaults < config file < environment < flags.
- Use `flag.Visit` to apply only the flags the user actually set.
- Keep config files in `os.UserConfigDir()`, optional by default, strict about unknown keys.
- Track where each setting came from, and say so in error messages.

## Further reading

- [Package flag: Visit](https://pkg.go.dev/flag#FlagSet.Visit)
- [os.UserConfigDir](https://pkg.go.dev/os#UserConfigDir)
- [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/latest/)
