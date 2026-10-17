# Project plan

This folder is the single place where we track what we are building, what is done, and why we made each decision. Nothing here is published on the website.

## Files

| File | What it tracks |
|---|---|
| [roadmap.md](roadmap.md) | Phases 0 to 5 and later work, with a checklist per phase. Content track and Kubernetes track side by side. |
| [curriculum.md](curriculum.md) | Every module and lesson, each with a checkbox. This is the master list of content. |
| [lesson-checklist.md](lesson-checklist.md) | Definition of done for a single lesson. Every lesson must pass it before its box in `curriculum.md` is ticked. |
| [case-studies.md](case-studies.md) | Candidate company stories and build-along projects, with research and sourcing status. |
| [k8s-track.md](k8s-track.md) | Step-by-step Kubernetes learning track for hosting this site. |
| [go-version-notes.md](go-version-notes.md) | Which Go release introduced each feature the course teaches. |
| [decisions.md](decisions.md) | Short decision records: what we chose, the alternatives, and why. |
| [backlog.md](backlog.md) | Ideas we agreed to postpone (login, in-browser runner, quizzes and so on). |

## How to use the checklists

- `[ ]` not started, `[~]` in progress, `[x]` done and reviewed.
- A lesson is only `[x]` after it passes [lesson-checklist.md](lesson-checklist.md) and Dharshan has reviewed it.
- When a phase is complete, update the status table below.

## Status

| Phase | Status |
|---|---|
| 0. Foundations (repo, site, CI, template) | Done; site runs on Kubernetes locally (K8s track steps 1 to 3) |
| 1. Fundamentals (modules 1 to 5) | All 5 modules drafted, awaiting review |
| 2. Concurrency and standard library | Modules 6 to 8 drafted, awaiting review |
| 3. Testing, performance, internals | Modules 9 to 11 drafted, awaiting review |
| 4. Production Go | Modules 12 to 15 drafted, awaiting review |
| 5. Projects and case studies | Shortener milestones 1 to 4, ShortLink operator, shortctl, job queue, greet, wordfreq, ratelimit, crawler, miniredis done; 9 case studies written from primary sources |
| Later: login, progress, in-browser runner | Parked |

Last updated: 2026-10-16
