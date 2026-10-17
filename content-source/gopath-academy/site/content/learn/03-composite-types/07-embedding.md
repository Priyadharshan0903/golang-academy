---
title: "3.7 Embedding and composition"
linkTitle: "3.7 Embedding"
weight: 7
description: "How struct embedding promotes fields and methods, how name conflicts are resolved, and why embedding is composition rather than inheritance."
goVersion: "1.27"
---

## What you will learn

- How to embed one struct in another, and what **promotion** means
- That embedding promotes methods too, with `sync.Mutex` as the classic example
- How Go resolves name conflicts between embedded fields
- Why embedding is **not** inheritance, and what that means for your designs

## The idea

Go has no classes and no inheritance. To build bigger types from smaller ones, you **compose** them: put one struct inside another. Embedding is a shortcut for composition. When you include a type as a field **without a field name**, its fields and methods are **promoted**: you can use them as if they were declared on the outer type.

## Promoted fields

{{< example file="examples/m03/embedding/embedding_test.go" region="promote" run="go test -run Example_promotedFields -v ./examples/m03/embedding" >}}

An embedded field still exists as a normal field. Its name is the **type name** (`Person`), so `e.Person.Name` always works. `e.Name` is a shorthand the compiler expands to `e.Person.Name`.

In a literal you set the embedded value through its type name: `Employee{Person: Person{...}}`.

## Promoted methods

Methods of the embedded type are promoted too. Methods are the topic of Module 4, but you can already use one of the most common cases, embedding `sync.Mutex` to make a type lockable:

{{< example file="examples/m03/embedding/embedding_test.go" region="mutex" run="go test -run Example_promotedMethods -v ./examples/m03/embedding" >}}

The zero value of `Counter` is ready to use: an unlocked mutex and a zero count.

Be aware that embedding **exports** what it promotes. Here `Lock` and `Unlock` become part of `Counter`'s public API, so any caller can lock it. For types in a public API it is often better to use a named, unexported field (`mu sync.Mutex`) and keep locking internal.

## Name conflicts

If several embedded types have a field with the same name, the rule is about **depth**:

- A field declared directly on the outer struct (depth 0) wins over promoted ones.
- If two promoted fields with the same name are at the same depth, using the short name is a **compile error** (ambiguous selector). You must spell out the path.

{{< example file="examples/m03/embedding/embedding_test.go" region="conflict" run="go test -run Example_nameConflicts -v ./examples/m03/embedding" >}}

Ambiguity is only an error when you actually use the ambiguous name, so a conflict can sit unnoticed until someone writes `a.Name`.

## Embedding is not inheritance

This is the most important part of the lesson. An `Employee` **contains** a `Person`; it **is not** a `Person`:

{{< example file="examples/m03/embedding/embedding_test.go" region="notinheritance" run="go test -run Example_notInheritance -v ./examples/m03/embedding" >}}

Differences from inheritance in Java, C# or Python:

- **No subtype relationship.** You cannot pass an `Employee` where a `Person` is required. Go gets polymorphism from **interfaces** instead (Module 4), and an `Employee` can satisfy an interface through its promoted methods.
- **No virtual dispatch.** If a promoted method of `Person` calls another `Person` method, it always calls `Person`'s version, even if `Employee` declares a method with the same name. The embedded value does not know it is embedded.
- **No `super`.** You call the inner implementation explicitly: `e.Person.Describe()`.

## When to embed

Embed when you want the outer type to **genuinely offer the inner type's behavior** as part of its own API:

- Adding locking to a type (`sync.Mutex`, with the caveat above).
- Wrapping a type to override one or two methods while keeping the rest, a common trick with interfaces like `http.ResponseWriter` (lesson 12.2).
- Sharing common fields across many structs, such as timestamps on database models.

Use a **named field** when the inner type is an implementation detail. A `Service` that uses a database should have `db *sql.DB`, not an embedded `*sql.DB`, otherwise every database method becomes a method of your service.

## Common mistakes

### Building class hierarchies with embedding

`Animal` embedded in `Dog` embedded in `Puppy` imports inheritance thinking and gets none of its benefits. Use small interfaces and plain composition.

### Accidentally exporting methods

Embedding a type puts all of its exported methods into your type's API. Readers of your documentation see them, and callers can use them.

### Embedding a pointer that is never set

You can embed a pointer type (`*Logger`). If it is nil, calling a promoted method or reading a promoted field panics. Initialize it in a constructor.

### Expecting overridden methods to be called from the inner type

There is no virtual dispatch. If you need the inner logic to call back into the outer type, pass an interface or a function to it explicitly.

## Key takeaways

- An embedded field is a field named after its type, whose fields and methods are promoted.
- Fields at shallower depth win; same-depth conflicts are an error when used.
- Embedding is composition: no subtyping, no virtual dispatch, no `super`.
- Embed to expose behavior on purpose; use a named field for implementation details.

## Further reading

- [Effective Go: Embedding](https://go.dev/doc/effective_go#embedding)
- [Struct types](https://go.dev/ref/spec#Struct_types) and [Selectors](https://go.dev/ref/spec#Selectors) in the spec, which define promotion and depth
- [Go FAQ: Why is there no type inheritance?](https://go.dev/doc/faq#inheritance)
