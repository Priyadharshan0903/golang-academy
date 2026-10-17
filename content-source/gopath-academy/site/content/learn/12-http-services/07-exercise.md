---
title: "12.7 Exercise: a notes JSON API"
linkTitle: "12.7 Exercise: notes API"
weight: 7
description: "Build a small JSON CRUD API from a written specification with ServeMux patterns, correct status codes and headers, checked by table-driven tests."
goVersion: "1.27"
---

## The task

The store is written for you: a concurrency-safe, in-memory collection of notes. Your job is the HTTP layer, written to this specification:

| Request | Success | Errors |
|---|---|---|
| `GET /notes` | 200, JSON array of all notes (`[]` when empty) | |
| `POST /notes` with `{"title": "...", "body": "..."}` | 201, the created note, `Location: /notes/{id}` | 400 if the body is not a JSON object, 422 if `title` is empty |
| `GET /notes/{id}` | 200, the note | 400 if `id` is not a number, 404 if it does not exist |
| `DELETE /notes/{id}` | 204, no body | 400, 404 |
| anything else on these paths | | 405 |

{{< exercise dir="m12/notes" >}}
Implement `NewHandler` in `starter.go`. The tests check every row of the specification, including headers and the 405 responses.
{{< /exercise >}}

```bash
cd exercises
go test ./m12/notes
```

## Hints

<details>
<summary>Hint 1: the routes</summary>

Four patterns with methods: `"GET /notes"`, `"POST /notes"`, `"GET /notes/{id}"` and `"DELETE /notes/{id}"`. With methods in the patterns, ServeMux produces the 405 responses by itself (lesson 12.1).

</details>

<details>
<summary>Hint 2: an empty list</summary>

`json.Marshal` turns a nil slice into `null`, not `[]`. `Store.List` returns a non-nil empty slice, so encoding it directly gives `[]`. Keep it that way.

</details>

<details>
<summary>Hint 3: keep handlers short</summary>

Write two small helpers: `writeJSON(w, status, v)`, which sets `Content-Type` before `WriteHeader`, and `parseID(w, r)`, which writes the 400 itself and returns `(id, ok)`.

</details>

The reference solution is in `exercises/m12/notes/solution.go`.

## Going further

- Use the strict `DecodeJSON` from lesson 12.3, and return problem details for errors.
- Add `PUT /notes/{id}` to replace a note, returning 200 or 404.
- Wrap the handler with the request ID and logging middleware from lesson 12.2.
