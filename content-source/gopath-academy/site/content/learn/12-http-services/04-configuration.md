---
title: "12.4 Configuration and the twelve-factor app"
linkTitle: "12.4 Configuration"
weight: 4
description: "Load configuration from the environment into one typed struct, validate it at start-up with every error reported at once, read secrets from files, and keep secrets out of logs."
goVersion: "1.27"
---

## What you will learn

- The twelve-factor rule for configuration, and why it fits containers
- Loading config into a typed struct with defaults and validation
- Testing configuration without touching the real environment
- Reading secrets from files, as Kubernetes provides them
- Logging configuration without leaking secrets

## The idea

The **twelve-factor app** methodology (written at Heroku in 2011, and still the norm for cloud services) says: store configuration in the **environment**. The same build artifact, one container image, then runs in development, staging and production, with only environment variables differing. Kubernetes, Docker and every platform support this directly.

Configuration is anything that varies between deployments: addresses, URLs of other services, credentials, limits, log levels. It is **not** things like your route table, which belong in code.

## One struct, loaded once

{{< example file="examples/m12/config/config.go" region="config" >}}

The pattern:

- **One `Config` struct** with typed fields (`time.Duration`, `*url.URL`, `slog.Level`), not strings to be parsed later all over the code.
- **Defaults** for everything that has a sensible one; only truly deployment-specific values are required.
- **Validate at start-up** and report **every** problem at once with `errors.Join`. A service with bad configuration should refuse to start with a clear message, not fail on the first request an hour later.
- **`getenv` is a parameter.** `main` passes `os.Getenv`; tests pass a map. No test ever has to modify the process environment, and tests can run in parallel.
- **Load once, pass down.** No package reads environment variables on its own; they receive what they need as constructor arguments. That keeps dependencies visible and code testable.

{{< example file="examples/m12/config/config_test.go" region="example" run="go test -run ExampleLoad -v ./examples/m12/config" >}}

## Flags or environment variables?

Both are fine and many services accept both. Command-line flags (the `flag` package, Module 14) are self-documenting (`-help`) and natural for CLIs. Environment variables are natural for services in containers, where the platform sets them from manifests. A common approach: environment variables for services, flags for tools, and if you support both, flags win.

Avoid configuration **files** for the basics. They have to be built into the image or mounted, and they invite a second, hidden source of truth. Mount a file when the configuration is genuinely structured and large.

## Secrets

Passwords, API keys and database URLs with credentials are configuration too, with extra rules:

- Never commit them, never bake them into images.
- Prefer reading them from **files**. Kubernetes Secrets and Docker secrets are usually mounted as files, and file contents do not show up in `env` output, process listings, or crash reports that dump the environment.

A widely used convention is a `_FILE` suffix: `DATABASE_URL_FILE=/run/secrets/db-url` means "read the value from this file":

{{< example file="examples/m12/config/config.go" region="secret" >}}

## Logging the configuration safely

Logging the effective configuration at start-up is very useful when debugging deployments. Implementing `slog.LogValuer` controls exactly what is printed:

{{< example file="examples/m12/config/config.go" region="redact" >}}

`TestSecretFromFile` logs the whole config and checks that the password does not appear. The log shows only `config.database_url_set=true`. Writing that test is cheap insurance: a careless `slog.Any("config", cfg)` in the future still cannot leak the secret.

## Common mistakes

### Reading os.Getenv deep inside packages

Configuration becomes invisible and untestable. Read once in `main`, pass values down.

### Failing on the first error

Operators then fix one variable per deploy. Collect all errors with `errors.Join`.

### Secrets in environment variables printed at start-up

Redact with `LogValuer`, and prefer files.

### Silent defaults for required settings

A default `BASE_URL` of `localhost` in production produces wrong links quietly. Make settings that differ in every environment required, or at least log them at start-up.

## Key takeaways

- Store configuration in the environment; one image runs everywhere.
- Load into one typed struct with defaults; validate at start-up and report all errors at once.
- Inject `getenv` so tests use a map.
- Read secrets from files where possible (`NAME_FILE`), and redact them in logs with `slog.LogValuer`.

## Further reading

- [The Twelve-Factor App: Config](https://12factor.net/config)
- [Kubernetes: Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
