# GoPath V3 repository and performance plan

Status: Draft for user review. Implementation, repository creation, pushes, and deployment have not started under this plan.

Build a maintainable Go learning website that delivers each lesson separately, keeps interactive features, and supports contributions through one GitHub repository. Preserve the original app and GoPath V2 for comparison. Keep explanations concise and practical; this refactor must preserve the imported curriculum.

## 1. Scope and version boundaries

- Prepare V3 in a separate working directory, with a proposed local preview on port 4175. Original port 4173 and V2 port 4174 remain available.
- Preserve both existing deployments and the supplied Downloads/gopath-academy folder.
- Keep the current React/Vinext hosting integration initially. Verify its route generation capabilities before choosing the page rendering implementation; a framework migration requires a revised plan.
- Keep the resource free, downloadable, and intended for developers learning Go.
- Exclude new curriculum writing, accounts, payments, a database, and a visual redesign from this refactor.
- Documents and code in the imported course remain source material. Build conversion must not execute their embedded instructions or arbitrary code.

## 2. Measured baseline

These are generated app JavaScript chunks, not total page downloads. Gzip sizes were estimated locally. Shared framework assets add to the browser payload.

| Build | App chunk bytes | Estimated gzip bytes |
|---|---:|---:|
| Original | 177,518 | 58,351 |
| GoPath V2 | 2,988,625 | 758,556 |

V2 imports a 2,819,403-byte generated JSON file into a client component. It contains course text, search data, exercises, embeds, and the complete source browser. The first load therefore includes material a learner has not opened.

Baseline inventory: 116 lessons, 15 modules, 27 exercises, nine case studies, eight project directories, 425 resolved code embeds, and 445 browsable source files. The import snapshot contains 619 files with recorded hashes. Preserve these inventories through migration.

## 3. Proposed repository

Use one repository and one canonical copy of the curriculum. The following is a target layout; map existing paths before moving files so Go modules and source embeds keep working.

```text
gopath-academy/
  web/                    # Routes, UI, styles, and runner endpoint
  content/
    lessons/
    modules/
    case-studies/
  examples/               # Actual Go source referenced by lessons
  exercises/              # Starters, tests, and solutions
  projects/               # Project documentation and Go implementations
  scripts/                # Content checks, generation, and packaging
  docs/                   # Plan, checklist, architecture, contribution guide
  .github/workflows/      # Pull request verification
  README.md
  LICENSE-CONTENT
  LICENSE-CODE
  THIRD-PARTY-NOTICES.md
```

- Keep Markdown/front matter and real Go files as editable sources. Preserve Go module boundaries, build tags, relative links, named regions, and source attribution.
- Remove the need for a second checked-in content snapshot. Retain the import manifest and a migration map for provenance; version history and release archives preserve old versions.
- Generate page data, search indexes, and ZIP downloads reproducibly. Ignore generated outputs, installed dependencies, caches, and local runtime files in Git.
- Split the large application component into a shell, course navigation, lesson reader, practice controls, project/source browser, search, and progress utilities. Format source for readable reviews.
- Give content CC BY-SA 4.0 and code MIT notices their proper scope; preserve other licenses and attribution already present.

## 4. Page and asset delivery

| Item | Proposed delivery |
|---|---|
| Lesson, module, case study, project prose | Build-generated HTML at a dedicated URL |
| Navigation | Small manifest of IDs, titles, ordering, and URLs |
| Progress, quiz, editor, copy/run controls | Focused interactive components |
| Editor and execution controls | Load when the learner opens the runnable example |
| Source browser | Load a file list first, then the selected file contents |
| Search | Load a separately generated index on first search use |
| Complete course and source ZIPs | Download on explicit request |

Use stable routes such as `/learn/<module>/<lesson>` and `/case-studies/<slug>`. Direct links, refreshes, headings, browser history, page titles, and descriptions must work. Preserve legacy hash links through a compatibility redirect or resolver.

Generate HTML for the full course during the build without importing the full course into browser JavaScript. Keep parsing and shortcode conversion in the build pipeline; preserve safe Markdown behavior. Do not preload all lesson bodies, source files, or the search index through automatic route prefetching.

Keep the existing runner behavior and truthful local-run instructions for unsupported examples. A purely static hosting destination would still need a separate execution endpoint; retain the current worker integration for the proposed comparison build.

## 5. Performance acceptance criteria

- Proposed budget: at most **250,000 gzip bytes of JavaScript** on a fresh overview or normal lesson load, including required shared framework chunks and automatic preloads. The editor and search have separate activation costs.
- No full-course JSON or complete source archive in the initial browser module graph or initial HTML payload.
- Opening a lesson retrieves its own content and shared navigation, without downloading all lessons.
- Source files, search data, and downloads are requested only when their corresponding features are used.
- Report raw, gzip, and actual transferred sizes separately. Report HTML/CSS and deferred feature costs alongside initial JavaScript.
- Measure a clean production build, avoiding stale chunks from older builds. Add an automated initial asset budget check.
- Record desktop and throttled mobile browser results for overview, a typical lesson, a long lesson, search, and editor activation. Record the setup with the results; do not claim speed improvements from file sizes alone.

The JavaScript budget is an engineering target, not an achieved result. If the retained framework prevents it, document the measured cause and propose a specific adjustment before relaxing the target or replacing the framework.

## 6. Implementation sequence after approval

| Phase | Work | Evidence required |
|---|---|---|
| 1. Baseline and routing proof | Create separate V3 workspace; inventory sources and routes; test one lesson rendered at a real URL with isolated interactivity | Repeatable baseline, direct-load HTML, chunk inspection, framework compatibility decision |
| 2. Canonical content migration | Establish source layout; resolve front matter, embeds, shortcodes, and links at build time | Inventory parity, migration map, no missing regions or unresolved content links |
| 3. Application refactor | Split UI; generate all content routes; defer source/search/editor; retain runner and downloads | Complete navigation and feature checks; initial bundle meets budget |
| 4. Learning state and compatibility | Preserve lesson IDs and ordering; support legacy URLs; validate progress behavior | Resume/completion/draft tests and legacy link checks |
| 5. Repository preparation | Remove unused starter code/dependencies where verified safe; document setup/licenses; add CI and release workflow | Reproducible fresh setup, passing required checks, reviewable local Git changes |
| 6. Comparison review | Start V3 on port 4175 and prepare a separate hosted comparison deployment under the available access policy | Three-version comparison record, browser evidence, final verification checklist |
| 7. GitHub publication | Resolve destination owner/name and inspect any existing repository; present final files and visibility for confirmation | Confirmed destination and permission to create/push, followed by verified repository URL |

Keep lesson IDs stable. Existing V2 progress remains untouched. A new hostname cannot read another hostname's browser storage, so provide validated JSON progress export/import for a separate V3 origin. When restoring drafts, display their content as text. State transfer must be explicit and must not overwrite existing progress without a learner choice.

## 7. Verification and GitHub workflow

- Content checks: all inventory entries, code regions, Markdown features, internal routes, heading links, licenses, and attribution survive migration.
- Application checks: build/type checking, representative UI flows, every generated content route, search, progress, legacy URLs, source browsing, runner success/error handling, and downloadable ZIP integrity.
- Go checks: run complete runnable package tests within supported module/toolchain scopes. Exercise starters may intentionally fail; verify reference solutions separately. Label external-service and toolchain-dependent checks clearly.
- CI: pin the chosen Node and Go toolchains, install from lockfiles, run content/application/Go checks, and fail the build on the initial JavaScript budget. Preserve environment-specific exclusions with reasons.
- Browser review: desktop/mobile layout, keyboard navigation, code overflow, clipboard/download behavior, and measured network requests. DOM tests alone do not satisfy this gate.
- GitHub documentation: quick start, content contribution guide, source-embed format, verification commands, license scopes, and version comparison links.
- Versioning: use release tags for original, V2, and V3 baselines where their actual history is available. Do not overwrite existing tags or manufacture shared ancestry. Keep existing comparison deployments while maintaining V3 going forward.

Repository visibility is proposed as public to match the free-resource goal. Confirm the final owner/name and inspect staged files, history, license scope, and repository-local credential references before a public push. Never copy credentials into source or CI examples.

Current comparison hosting is private under the connected workspace policy. Public GitHub availability does not make that deployment public. Choose a public-capable hosting destination separately if required; do not change existing audiences to complete this refactor.

## 8. Completion and review

Deliver V3 source, a reproducible build, verified downloadable course/source archives, a measured before/after bundle report, a three-version comparison record, and the completed checklist. Retain links to this plan and the checklist in the future repository documentation.

Planning is the only authorized work in the current turn. After the user approves this plan, begin phases 1–6. GitHub creation/push waits for the concrete destination and final publication confirmation in phase 7.

Review points: target repository layout, separate V3 comparison version, 250 KB initial JavaScript budget, and staged GitHub publication.
