---
title: "9. Testing in depth"
weight: 9
description: "Table-driven tests, helpers and golden files, testable examples, fakes, httptest, fuzzing, benchmarks with b.Loop, coverage, integration tests and testing/synctest."
---

Testing is built into Go: `go test`, the `testing` package and a handful of conventions cover almost everything, with no framework to install. This module goes from the basics you have already used to fuzzing, benchmarks and deterministic concurrency tests.

**Time:** about 4 hours, plus exercises. One exercise turns things around: you write the tests, and the repository checks whether they catch deliberately planted bugs.

{{< cards >}}
  {{< card link="01-table-driven" title="9.1 Table-driven tests and subtests" subtitle="The standard Go test shape, and testing your tests." >}}
  {{< card link="02-helpers" title="9.2 Helpers, cleanup and golden files" subtitle="t.Helper, t.TempDir, t.Setenv, t.Context." >}}
  {{< card link="03-examples" title="9.3 Testable examples" subtitle="Documentation that is checked by go test." >}}
  {{< card link="04-fakes" title="9.4 Mocks, fakes and interfaces" subtitle="Test doubles without a mocking framework." >}}
  {{< card link="05-httptest" title="9.5 httptest" subtitle="Testing handlers and clients." >}}
  {{< card link="06-fuzzing" title="9.6 Fuzzing" subtitle="Let the computer find your edge cases." >}}
  {{< card link="07-benchmarks" title="9.7 Benchmarks with b.Loop" subtitle="Measuring speed and allocations." >}}
  {{< card link="08-coverage" title="9.8 Coverage and integration tests" subtitle="-cover, -short, TestMain, build tags." >}}
  {{< card link="09-synctest" title="9.9 Testing concurrent code with synctest" subtitle="Fake time, no flaky sleeps." >}}
{{< /cards >}}
