---
title: "13.3 Migrations"
linkTitle: "13.3 Migrations"
weight: 3
description: "Version your database schema with migrations: embed SQL files in the binary, apply them in a transaction under an advisory lock so replicas can start together, and change schemas without downtime using expand and contract."
goVersion: "1.27"
---

## What you will learn

- What migrations are and why schema changes must be versioned
- A complete, small migrator: embedded SQL files, a version table, a transaction and an advisory lock
- Testing that migrations are idempotent and safe to run concurrently
- Changing a schema without downtime: **expand and contract**
- When to run migrations, and which tools to use

## The idea

Your code expects a particular database schema: these tables, these columns. As the code changes, so must the schema, in every environment, in the same order. A **migration** is one numbered schema change, kept in version control next to the code. The database records which migrations it has applied, and a migrator applies the missing ones in order.

## A small migrator

The shortener (lesson 13.7) has its own migrator, short enough to read in full. The migrations are plain SQL files:

{{< example file="projects/shortener/internal/postgres/migrations/0001_create_links.sql" >}}

And the code that applies them:

{{< example file="projects/shortener/internal/postgres/migrate.go" >}}

Every part has a job:

- **`//go:embed migrations/*.sql`** compiles the SQL into the binary (lesson 8.5). A deployment is one file, and version 2.3 of the program always carries exactly the migrations of version 2.3.
- **`schema_migrations`** records the applied versions. Files are applied in name order, which is why they start with a zero-padded number.
- **One transaction.** PostgreSQL can roll back schema changes (DDL), so either all pending migrations apply or none do. A half-applied migration is the worst state to debug at 3 a.m.
- **`pg_advisory_xact_lock`** takes a lock identified by a number, held until the transaction ends. When three replicas of a service start at the same moment, the first takes the lock and migrates; the others wait, then find nothing left to do.

The tests check exactly those properties, against a real database:

```text
store_test.go:33: applied [0001_create_links.sql], then nothing
```

`TestMigrateIsIdempotent` runs it twice; `TestMigrateConcurrently` runs it from five goroutines at once and checks every migration was applied exactly once.

A rule that keeps migrations sane: **never edit a migration that has been applied anywhere** outside your laptop. Fix mistakes with a new migration. Applied migrations are history.

A limitation of running everything in one transaction: a few statements cannot run inside a transaction, notably `CREATE INDEX CONCURRENTLY`, which builds an index without blocking writes. Migration tools let you mark such migrations as non-transactional.

## Changing schemas without downtime

During a rolling deployment, old and new versions of your service run **at the same time**, against the same database. Every migration must therefore work with both the code before it and the code after it. That rules out a single-step rename:

```sql
ALTER TABLE links RENAME COLUMN url TO target;  -- old replicas break immediately
```

Instead, **expand and contract**, over several deployments:

1. **Expand.** Add the new column (`target`). Deploy code that writes **both** columns and reads the old one.
2. **Migrate data.** Backfill `target` from `url` in batches.
3. **Switch.** Deploy code that reads `target` (still writing both).
4. **Contract.** Once no running code uses `url`, stop writing it, and in a later migration drop it.

It is slower, but no step ever breaks a running version. The same thinking applies to adding `NOT NULL` columns (add with a default, or nullable first), and to dropping anything.

Also watch for **locks**: some `ALTER TABLE` forms rewrite or lock the whole table, blocking all queries for the duration. Set a `lock_timeout` in migrations so a blocked migration fails fast instead of queueing every request behind it.

## When to run migrations

- **At start-up**, as the shortener does by default (`MIGRATE=true`). Simple, and the advisory lock makes it safe with many replicas. The cost: a slow migration delays start-up, and the service's database user needs permission to change the schema.
- **As a separate step** before rolling out the new version: a Kubernetes Job, an init container, or a CI/CD stage running the same binary with a `migrate` command. This gives you more control, and lets the service run with a user that cannot alter tables. Larger teams usually end up here.

## Tools

The migrator above is about 60 lines because PostgreSQL does the heavy lifting. Established tools add down migrations, non-transactional migrations, CLIs and multi-database support:

- **goose** (`github.com/pressly/goose`): SQL or Go migrations, embeddable, popular in Go services.
- **golang-migrate** (`github.com/golang-migrate/migrate`): many databases and sources, with a CLI.
- **Atlas** (`ariga.io/atlas`): declarative: you describe the desired schema and it plans the changes.

Whichever you use, the principles are the same: versioned files in the repository, applied in order, recorded in the database, never edited after release.

## Common mistakes

### Editing applied migrations

Environments diverge silently. Add a new migration instead.

### Breaking changes in one step

Renames and drops break the old version during a rolling deploy. Expand, then contract.

### Long-locking migrations at peak traffic

Know which `ALTER TABLE` forms lock or rewrite tables; set `lock_timeout`; build indexes concurrently.

## Key takeaways

- Migrations are numbered, versioned schema changes applied in order and recorded in the database.
- Embed them in the binary; apply them in one transaction under an advisory lock so concurrent replicas are safe.
- Never edit an applied migration.
- For zero downtime, every migration must work with both the old and new code: expand, migrate, switch, contract.

## Further reading

- [PostgreSQL: advisory locks](https://www.postgresql.org/docs/current/explicit-locking.html#ADVISORY-LOCKS)
- [goose](https://github.com/pressly/goose) and [golang-migrate](https://github.com/golang-migrate/migrate)
