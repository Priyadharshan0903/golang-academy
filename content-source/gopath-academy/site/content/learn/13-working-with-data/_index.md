---
title: "13. Working with data"
weight: 13
description: "PostgreSQL from Go with database/sql and pgx, connection pools, migrations, transactions and locking, caching with Redis, and two projects: the shortener on Postgres and a job queue."
---

Most services are, at their core, a careful layer over a database. This module covers the parts that decide whether that layer is correct under load: pools, parameters, transactions, locks, migrations and caches. It uses **PostgreSQL** and **Redis**, the most common pair in Go services, and every example runs against real servers.

The database examples live in the `examples/thirdparty` module (they need the pgx and go-redis libraries). They are skipped unless a database is available; the [setup takes two docker commands](https://github.com/Priyadharshan0903/gopath-academy/tree/main/examples/thirdparty#databases-module-13).

**Time:** about 5 to 6 hours, plus the exercise and projects.

{{< cards >}}
  {{< card link="01-database-sql" title="13.1 database/sql and pooling" subtitle="Queries, NULLs, injection, pool limits." >}}
  {{< card link="02-pgx" title="13.2 pgx and PostgreSQL" subtitle="Structs, error codes, batch and COPY." >}}
  {{< card link="03-migrations" title="13.3 Migrations" subtitle="Versioned, embedded, safe with many replicas." >}}
  {{< card link="04-transactions" title="13.4 Transactions and context" subtitle="Lost updates, locks, deadlocks, cancellation." >}}
  {{< card link="05-caching-redis" title="13.5 Caching with Redis" subtitle="Cache-aside, TTLs, stampedes, failure." >}}
  {{< card link="06-exercise" title="13.6 Exercise: a read-through cache" subtitle="Expiry and request coalescing by hand." >}}
  {{< card link="07-project-shortener-postgres" title="13.7 Project: shortener on Postgres" subtitle="Milestone 3, with a store contract test." >}}
  {{< card link="08-project-jobqueue" title="13.8 Project: a Postgres job queue" subtitle="SKIP LOCKED, retries, crash recovery." >}}
{{< /cards >}}
