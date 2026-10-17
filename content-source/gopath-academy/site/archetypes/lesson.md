---
title: "{{ replace .File.ContentBaseName "-" " " | strings.TrimLeft "0123456789 " | title }}"
weight: 10
description: "One sentence on what this lesson teaches."
goVersion: "1.27"
---

<!--
Template notes (delete this comment when done):
- Put example code in examples/mNN/<topic>/ with tests, then turn the
  placeholders below into real shortcodes by removing the /* and */ markers.
- Follow plan/lesson-checklist.md before marking the lesson done.
-->

## What you will learn

- 
- 
- 

## The idea

Explain the concept in plain words before showing any code.

## Why it matters

A real situation where this concept changes how you write code.

## A minimal example

{{</* example file="examples/mNN/topic/file.go" region="minimal" */>}}

## Going deeper

{{</* example file="examples/mNN/topic/file.go" region="deeper" */>}}

## Common mistakes

### Mistake one

### Mistake two

## Exercise

{{</* exercise dir="mNN/topic" */>}}
Describe what to implement.
{{</* /exercise */>}}

## Key takeaways

- 
- 
- 

## Further reading

- 
