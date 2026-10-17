# V1 and V2 website separation plan

Status: Implemented and deployed on 17 October 2026. Both existing websites now have isolated builds and verified learner downloads. Real-browser verification remains pending because macOS blocks Chromium launch; see the checklist and release records.

Run Go by Doing V1 and GoPath V2 as two independently deployable websites. Give each site a download containing its own learning material and runnable code. Keep the website application, publishing configuration, and internal planning files out of those learner downloads.

Track work in [V1 and V2 separation checklist](V1-V2-SPLIT-CHECKLIST.md). This plan is separate from the existing [V3 repository and performance plan](GOPATH-V3-REFACTOR-PLAN.md); V3 remains future work.

## 1. Current boundaries

These findings come from the current checkout, not a fresh audit of either hosted website.

| Area | V1 | V2 |
| --- | --- | --- |
| App component | `app/learning-app.tsx` | `app/gopath-app.tsx` |
| Content | `app/data/curriculum.ts`, `cases.ts`, project guide and project files | `content-source/gopath-academy`, compiled into `app/data/gopath.json` |
| Curriculum inventory | App describes 48 lessons, 16 modules, four teaching scenarios, and a guided website-checker | Generated data contains 116 lessons, 15 modules, 27 exercises, nine case studies, eight projects, and 445 browsable source files |
| Progress key | `go-by-doing.progress.v1` | `go-by-doing.gopath-v2.progress.v1` |
| Download generation | `scripts/generate-learning-assets.mjs` and `generate-project-assets.mjs` | ZIP creation is currently part of `scripts/import-gopath.mjs` |

`app/page.tsx` currently serves V2. This checkout contains one hosting manifest and Site identity; README and comparison links describe a separate original deployment. Verify both live destinations before creating another Site.

Only the V2 ZIP is currently present in `public/downloads`. It contains all 619 files in the imported snapshot, including Hugo website files, CI workflows, internal plans, and deployment infrastructure. V1's retained generators already export lesson notes and runnable examples, but their outputs need regeneration and validation in the V1 build.

## 2. Recommended structure

Keep one source repository with two app build roots. Each website gets its own entry point, metadata, public assets, download output, hosting identity, build command, and release record.

```text
go-by-doing/
  apps/
    v1/                       # Original website and its hosting manifest
    v2/                       # GoPath website and its hosting manifest
  content/
    v1/                       # Original curriculum and project sources
    v2/                       # GoPath curriculum and project sources
  scripts/
    content/                  # Content validation and format conversion
    downloads/                # Separate V1 and V2 learner ZIP exporters
    release/                  # Build and publication preparation
  docs/
    review-plans/
```

The tree is a target, not an instruction to move files immediately. First map all imports, relative paths, named code regions, Go modules, and generated files. Preserve the hashed V2 import snapshot during extraction; do not invalidate the existing provenance check by editing it in place.

Share UI primitives only where both apps already need the same behavior. Keep curriculum manifests, search indexes, progress, drafts, project checkpoints, and version-specific rendering separate. Avoid a generic curriculum engine in this first split.

Prove the hosting workflow can publish each app from its build root. If it requires a standalone project root, prepare an isolated publication checkout containing that app, its required content, resolved shared dependencies, and its own hosting manifest. Verify this with one representative page before completing the move. A repository split is optional later; two websites do not require two repositories.

## 3. Two website contracts

| Contract | V1 website | V2 website |
| --- | --- | --- |
| Identity | Go by Doing, original curriculum | GoPath Edition, deeper curriculum |
| Curriculum | Original lessons, quizzes, teaching scenarios, guided project | GoPath lessons, exercises, case studies, projects, source browser |
| Main download | `go-by-doing-v1-content.zip` | `gopath-v2-content.zip` |
| State | Preserve V1 keys and IDs | Preserve V2 keys and IDs |
| Release | Build, deploy, and roll back independently | Build, deploy, and roll back independently |

Reuse the existing V1 and V2 hostnames and Site identities where verified. Configure the other edition's URL in one place per app. Replace local-port comparison links in production, and make each homepage explain which edition the learner is using.

Keep existing lesson URLs, hash routes, heading links, downloads, and browser state working. Existing download URLs should resolve to the new learner archive or redirect to it. Keep V1 project starter and solution downloads available.

Local storage belongs to an origin. Keeping an existing hostname preserves access to its stored state; changing hostname does not transfer that state. If an origin must change, provide validated progress/draft export and import, with explicit conflict handling. Do not merge V1 and V2 progress or treat matching lesson titles as matching IDs.

Retain the code runner endpoint or an explicitly configured equivalent on each site. Verify that a V1 build does not contain V2 curriculum data and vice versa. Separate builds must also exclude the other edition's public downloads, search data, and source-browser content.

## 4. Content-only download contract

The default ZIP is a learner pack: readable lesson text plus code needed to follow the course. Learners should not need Node.js, React, Vinext, or Hugo to read it. Go examples retain their actual Go requirements; database and cluster exercises retain their documented service requirements.

Include:

- A start-here README, ordered curriculum index, local commands, and toolchain requirements.
- Standard Markdown lessons, case studies, project explanations, citations, and required diagrams.
- Runnable Go examples, expected outputs, practice tasks, quizzes, hints, and explained answers where present.
- Exercise starters, tests, reference solutions, project implementations, and required configuration examples.
- Every required `go.mod`, `go.sum`, embedded asset, fixture, SQL migration, and nested-module dependency file.
- License notices, attribution, edition/release metadata, and a file manifest with checksums.

Exclude website UI source, application APIs, website dependencies and lockfiles, hosting credentials/configuration, CI workflows for publishing the website, Hugo templates/configuration, internal plans, caches, build output, and editor/runtime state. Go dependency files remain included.

Select files through an explicit exporter allowlist and dependency checks. A folder called `deploy` is not automatically excluded: learner-facing Docker/Kubernetes examples and referenced Go server code may be part of a lesson. Document why each such file is included. Exclude infrastructure used only to publish this website.

Keep each Go module's internal layout intact. Preserve build tags and starter/solution behavior instead of relocating solution files into a layout that breaks existing commands. Reference solutions remain clearly labeled; intentionally failing starter tests are not release failures.

Convert Hugo shortcodes and internal `gopath-*` renderer tokens into ordinary Markdown, code blocks, callouts, and exercise sections. Resolve code-region embeds into visible code. Rewrite internal lesson links and local commands to match the exported layout. For example, the word-frequency checkpoint currently reads `../../site/content/learn/*/*.md`; the exported command must point to files that actually exist in the learner pack.

An illustrative archive layout is:

```text
<edition>-content/
  README.md
  INDEX.md
  lessons/
  case-studies/
  examples/
  exercises/
  projects/
  <required learner deployment examples>
  go.mod                     # Where required by this edition
  <applicable license notices>
  manifest.json
```

Adapt the final layout to verified module boundaries and references. Keep website-source downloads separate if offered later; they are not the primary learning download.

## 5. Implementation sequence

| Phase | Work | Completion evidence |
| --- | --- | --- |
| 1. Inventory and build proof | Map both apps, source ownership, routes, state, download dependencies, and actual hosting identities. Prove two isolated build/publication roots. | Inventory, dependency map, destination map, and independent build proof |
| 2. App separation | Extract V1 and V2 app roots with correct content, assets, metadata, state keys, runner configuration, and edition links. | Each app runs without the other's curriculum; route and state checks pass |
| 3. Learner exporters | Generate edition-specific Markdown and runnable-code ZIPs from canonical sources; validate dependency closure and licenses. | Two reproducible archives, file manifests, and successful clean-unzip checks |
| 4. Download experience | Update header, reader, exercise, project, source-browser, and Resources downloads. Preserve old download URLs and project-specific archives. | Every visible download points to the correct edition and verified contents |
| 5. Verification and release | Verify UI, routes, state, runner, search, downloads, and independent deployment/rollback. Publish each app to its confirmed destination. | Two verified website URLs, release records, archive checksums, and completed release checklist |

Start with inventory and the two-build proof. Keep the current working websites available during migration. Do not fold the V3 rendering/performance refactor or a new visual redesign into this separation.

## 6. Verification and release

Run content completeness and link checks, TypeScript checks, production builds, and representative browser flows for both apps. Check desktop/mobile and light/dark themes, keyboard navigation, completion/resume, drafts, search, lesson navigation, copy/run, and download behavior. Inspect initial app assets to confirm curriculum isolation.

Unzip each learner archive into a clean directory. Check readable Markdown without website tooling, relative links, all referenced assets, individual example commands, supported Go package tests, and reference exercise solutions. Record external-service and toolchain-dependent checks separately from passed checks. Verify the V1 website-checker starter and solution independently.

Build archives from a fresh staging directory, with sorted entries and normalized metadata, so removed files cannot remain in later ZIPs and identical inputs produce identical archives. Record edition, release identifier, source revision, content inventory, per-file hashes, archive checksum, and verification scope. Generate the website and its matching download from the same content revision.

Before publishing, read each Site's current identity and audience; preserve its existing access. Public learner access, if desired, requires a hosting destination that supports that audience. Reuse an existing deployment rather than registering a duplicate. Release one app at a time, verify its download, and confirm the other site still works. Record how to roll back each site together with its corresponding download.

## 7. Decisions to resolve during implementation

- Final public names and hostnames; confirm which existing deployments can be reused.
- Whether current origins can remain, or state export/import is required.
- Hosting support for two app roots versus isolated publication checkouts.
- Exact learner deployment examples and assets required by the course dependency audit.
- Whether solutions stay in the default pack or receive an additional separate download. The default proposal retains and clearly labels them to preserve current learning behavior.

These decisions do not block creating the plan or beginning the inventory. Resolve them before their dependent migration or release work.

## 8. Definition of done

V1 and V2 have independently buildable and deployable apps, correct version-specific content, working edition links, preserved state behavior, and verified destinations. Each website downloads its own complete learner pack, with readable lessons and runnable examples, without website publishing machinery. Existing links have a compatibility path, each release has verification evidence and a rollback reference, and the tracking checklist records remaining limitations honestly.
