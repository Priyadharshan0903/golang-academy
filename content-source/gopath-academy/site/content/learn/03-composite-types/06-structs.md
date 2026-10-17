---
title: "3.6 Named types, structs, tags and anonymous structs"
linkTitle: "3.6 Named types and structs"
weight: 6
description: "Defining your own types, building structs with literals, struct comparison and copying, struct tags, and anonymous structs for one-off data."
goVersion: "1.27"
---

## What you will learn

- How `type` creates a new, distinct type, and why that prevents bugs
- How to declare structs and build them with literals
- How copying and comparison work for structs
- What struct tags are and who reads them
- When anonymous structs are handy

## Named types

`type Name Underlying` creates a **new type**. It has the same representation as its underlying type, but the compiler treats it as different:

{{< example file="examples/m03/structs/structs_test.go" region="named" run="go test -run Example_namedTypes -v ./examples/m03/structs" >}}

`Celsius` and `Fahrenheit` are both `float64` underneath, but you cannot pass one where the other is expected without an explicit conversion. Named types let the compiler catch unit mix-ups, IDs of the wrong kind (`UserID` vs `OrderID`), and enums like `Weekday` from lesson 2.3. And, as you will see in Module 4, **you can define methods on named types**.

There is also a type **alias**, written with `=`: `type Temp = Celsius`. An alias is just another name for the **same** type, with no conversion needed. Aliases exist mainly for moving types between packages during refactors. When you want a new type, leave out the `=`.

## Structs

A struct groups named fields. It is Go's main way to model data, playing the role that classes play elsewhere, minus inheritance.

{{< example file="examples/m03/structs/structs_test.go" region="literal" run="go test -run Example_literals -v ./examples/m03/structs" >}}

Points worth knowing:

- **Use field names in literals** (`User{ID: 1, Name: "Ana"}`). Positional literals (`User{1, "Ana", "", ...}`) break as soon as someone adds a field, and `go vet` warns about them for structs from other packages.
- **Omitted fields get their zero values.** That is why it pays to design structs so the zero value of every field is a sensible default.
- **`%+v`** in `fmt` prints field names, which is the quickest way to inspect a struct while debugging. `%#v` prints it as Go syntax.
- Exported field names (uppercase) are visible from other packages, and they are the only fields that packages like `encoding/json` can see.

### Pointers to structs

You will often create structs through a pointer: `u := &User{Name: "Ana"}`. Field access is the same (`u.Name`), as lesson 2.8 showed. Functions that build structs usually return a pointer when the struct is large or will be modified, and a value when it is small and immutable.

## Copying and comparing

{{< example file="examples/m03/structs/structs_test.go" region="copy" run="go test -run Example_copyAndCompare -v ./examples/m03/structs" >}}

- Assignment copies every field. But a **shallow** copy only: if a field is a slice, map or pointer, both copies share what it points to.
- Structs are comparable with `==` when all their fields are comparable. A struct containing a slice or map is not comparable, and cannot be a map key.

## Struct tags

A **tag** is a string literal after a field's type. The compiler ignores it; libraries read it at run time with reflection:

{{< example file="examples/m03/structs/structs_test.go" region="tags" run="go test -run Example_tags -v ./examples/m03/structs" >}}

The convention is space-separated `key:"value"` pairs. You will meet tags constantly:

| Tag key | Read by |
|---|---|
| `json` | `encoding/json` (field name, `omitempty`, `-` to skip) |
| `xml`, `yaml`, `toml` | the matching encoders |
| `db` | SQL helper libraries such as sqlx |
| `validate` | validation libraries |
| `env` | config loaders |

`go vet` checks that tags are well formed. A typo like `json:name` (missing quotes) is silently useless otherwise. The JSON details are covered in lesson 8.2.

## Anonymous structs

A struct type does not need a name. Anonymous structs are perfect for data that is used in exactly one place:

{{< example file="examples/m03/structs/structs_test.go" region="anon" run="go test -run Example_anonymousStructs -v ./examples/m03/structs" >}}

Their most common use is **table-driven tests**, which you have already seen in the exercise tests. They are also handy for decoding one-off JSON responses and for template data.

## The empty struct

`struct{}` has no fields and takes **zero bytes**. You saw it as a set value in lesson 3.5. It is also used as a signal on channels (`chan struct{}`) in Module 6, where only the event matters, not any data.

## Common mistakes

### Positional struct literals

`Point{1, 2}` is fine for tiny, stable types. For anything else, name the fields.

### Expecting deep copies

Copying a struct with a slice field gives you two structs sharing one slice. Clone the slice when you need independence.

### Unexported fields and encoders

`encoding/json` skips unexported fields silently. If a field is missing from your JSON output, check its first letter.

### Malformed tags

`json: "name"` (with a space) or `json:name` are ignored. Let `go vet` check them.

## Key takeaways

- `type T U` creates a distinct type; use named types to prevent unit and ID mix-ups. `type T = U` is only an alias.
- Build structs with field-named literals; omitted fields are zero.
- Struct assignment is a shallow copy; `==` works if every field is comparable.
- Tags are metadata for libraries such as `encoding/json`, and `go vet` checks their format.
- Anonymous structs are ideal for test tables and one-off data.

## Further reading

- [Type declarations](https://go.dev/ref/spec#Type_declarations) and [Struct types](https://go.dev/ref/spec#Struct_types) in the spec
- [Effective Go: Composite literals](https://go.dev/doc/effective_go#composite_literals)
- [Package reflect: StructTag](https://pkg.go.dev/reflect#StructTag)
