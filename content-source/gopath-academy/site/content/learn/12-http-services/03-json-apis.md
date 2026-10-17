---
title: "12.3 Request validation and JSON APIs"
linkTitle: "12.3 JSON APIs"
weight: 3
description: "Decode request bodies strictly and safely, validate business rules separately, choose the right status codes, and return errors as RFC 9457 problem details, with every failure path tested."
goVersion: "1.27"
---

## What you will learn

- How to decode JSON request bodies safely: size limits, unknown fields, trailing data
- Turning decoding errors into helpful client messages
- Separating **decoding** from **validation**
- Which status codes to use: 400, 413, 415, 422 and friends
- Returning errors in the standard **problem details** format (RFC 9457)

## The idea

A JSON endpoint does three distinct things with a request body, and each can fail differently:

1. **Read** it: is it JSON, is it too big? → `415`, `413`
2. **Decode** it into a Go struct: is it well-formed, are the types right, are there unknown fields? → `400`
3. **Validate** it against business rules: is the email an email, is the age plausible? → `422`

Keeping these separate keeps handlers short and error messages precise.

## Decoding safely

`json.NewDecoder(r.Body).Decode(&v)` is how most tutorials decode. It has four problems in a public API: no size limit (a client can send gigabytes), unknown fields are silently ignored (a typo like `"emial"` is accepted and dropped), anything after the first value is ignored, and its error messages are not meant for clients. This helper fixes all four:

{{< example file="examples/m12/jsonapi/jsonapi.go" region="decode" >}}

- `mime.ParseMediaType` accepts `application/json; charset=utf-8` as well as the plain form.
- `http.MaxBytesReader` fails the read after `maxBytes`, and tells the server to close the connection rather than read the rest.
- `DisallowUnknownFields` rejects fields the struct does not have.
- The second `Decode` must hit `io.EOF`: the body holds exactly one JSON value.
- `decodeError` maps errors to messages a client can act on. It checks error **types** (`*json.SyntaxError`, `*json.UnmarshalTypeError`, `*http.MaxBytesError`) with `errors.As`, never message text.

That last rule matters more than ever: since Go 1.27, `encoding/json` is implemented on top of `encoding/json/v2` (lesson 8.2), and the wording of some error messages differs between versions. The example's test for unknown fields checks only the status and that the field name appears, so it passes on Go 1.24 and Go 1.27 alike.

## Validation is separate

{{< example file="examples/m12/jsonapi/jsonapi.go" region="handler" >}}

`Validate` is a method on the request type, returning one message per field. Nothing in it knows about HTTP, so it is easy to unit test, and the handler stays a straight line: decode, validate, act, respond. Validation libraries using struct tags exist, but for most services plain Go is clearer and checked by the compiler.

The response type `User` is a **different type** from the request type `CreateUser`. Separate types for input and output let them evolve independently and make it impossible to accidentally accept fields clients should not set (like `id`). Note `omitzero` (Go 1.24) on `Age`, which omits the field when it is zero.

## Problem details

For errors, use the standard format from **RFC 9457**, "Problem Details for HTTP APIs": a JSON object served as `application/problem+json` with `title`, `status` and an optional `detail`, extensible with your own members:

{{< example file="examples/m12/jsonapi/jsonapi.go" region="problem" >}}

Clients and API gateways recognize the format, and every error from your API has the same shape.

## Every path, tested

{{< example file="examples/m12/jsonapi/jsonapi_test.go" region="example" run="go test -run Example -v ./examples/m12/jsonapi" >}}

Each line is one request and its exact response. Note:

- A successful create returns **201 Created** with a `Location` header pointing at the new resource.
- The missing `age` in the second request is valid and omitted from the response.
- Validation reports **all** field errors at once, so a client can fix everything in one round trip. (Map keys are sorted in JSON output, so the output is stable.)
- A string where a number belongs is a **decoding** error (400), not a validation error.
- Two JSON objects in one body are rejected.
- The wrong `Content-Type` gets **415 Unsupported Media Type**.

## Choosing status codes

| Code | Use it when |
|---|---|
| 200 OK | a successful read or update that returns a body |
| 201 Created | a resource was created; set `Location` |
| 204 No Content | success with no body, typically DELETE |
| 400 Bad Request | the request is malformed: not JSON, wrong types, unknown fields |
| 401 Unauthorized | no or invalid credentials |
| 403 Forbidden | valid credentials, but not allowed |
| 404 Not Found | the resource does not exist |
| 405 Method Not Allowed | ServeMux sends it for you |
| 409 Conflict | the request conflicts with current state, such as a duplicate key |
| 413 Content Too Large | the body is over your limit |
| 415 Unsupported Media Type | wrong `Content-Type` |
| 422 Unprocessable Content | well-formed, but fails validation |
| 429 Too Many Requests | rate limited; set `Retry-After` |
| 500 Internal Server Error | a bug or a failed dependency; never show internal details |
| 503 Service Unavailable | temporarily unable, for example during shutdown |

Some teams use 400 for validation errors too. Either is defensible; be consistent and document it.

## Common mistakes

### Decoding without a size limit

One request can exhaust memory. Always use `http.MaxBytesReader` (or `http.MaxBytesHandler` for the whole server, lesson 12.6).

### Matching error strings

They change between Go versions. Use `errors.As` with the error types.

### Showing internal errors to clients

`"pq: connection refused to 10.0.3.7"` helps attackers and confuses users. Log the details; return a generic message and a request ID (lesson 12.8 does this).

### Writing headers after the body

Headers and the status must be set before the first `Write`. `WriteJSON` sets `Content-Type`, then `WriteHeader`, then writes.

## Key takeaways

- Decode strictly: check `Content-Type`, limit size, disallow unknown fields, require a single value.
- Classify decode errors by type with `errors.As`, never by message.
- Validate separately with plain Go methods that return per-field errors.
- Use distinct request and response types.
- Return errors as `application/problem+json` (RFC 9457) with accurate status codes.

## Further reading

- [RFC 9457: Problem Details for HTTP APIs](https://www.rfc-editor.org/rfc/rfc9457)
- [RFC 9110: HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110), the definitions of every status code
- [Package encoding/json](https://pkg.go.dev/encoding/json)
