# GoPath V3 repository and performance checklist

Status: Planning complete; awaiting user review. Unchecked items are future work, not verified results.

Plan: [GoPath V3 repository and performance plan](GOPATH-V3-REFACTOR-PLAN.md).

## Plan and approval

- [x] Record the current app-chunk measurements and their scope.
- [x] Define repository structure, content delivery, compatibility, and performance target.
- [x] Define implementation phases and required verification evidence.
- [x] Keep plan and checklist in separate files.
- [ ] User approves the plan or supplies revisions.

## Baseline and routing proof

- [ ] Create separate V3 workspace and preserve original/V2/source folder.
- [ ] Record baseline content inventory, stable IDs, and Go module boundaries.
- [ ] Capture clean-build initial assets and browser network behavior.
- [ ] Verify one real lesson URL delivers rendered content on direct request.
- [ ] Verify editor/progress isolation and framework hosting compatibility.

## Content and application

- [ ] Establish one canonical source tree and retain provenance/migration mapping.
- [ ] Preserve 116 lessons, 15 modules, 27 exercises, nine cases, and eight project directories.
- [ ] Preserve 425 code embeds and all 445 browsable source files.
- [ ] Validate source regions, supported Markdown/shortcodes, links, and heading anchors.
- [ ] Split and format application components; remove the whole-course client import.
- [ ] Generate dedicated lesson/module/case/project routes with meaningful titles.
- [ ] Load search, editor, and source bodies on demand; prevent bulk prefetching.
- [ ] Retain runnable examples, accurate local commands, and runner error handling.
- [ ] Preserve licenses/attribution and verify generated download contents.

## Compatibility and performance

- [ ] Stable IDs/order and legacy hash links work.
- [ ] Completion, resume, drafts, and validated progress export/import work.
- [ ] Existing app state remains untouched; restoration offers conflict handling.
- [ ] Fresh overview and lesson initial JavaScript each meet the 250,000-byte gzip budget.
- [ ] Initial HTML and JavaScript exclude the full course/source dataset.
- [ ] Capture raw/compressed/transferred sizes and deferred feature costs separately.
- [ ] Verify requests in a real browser; record desktop/mobile performance setup.
- [ ] Record desktop/mobile visual, keyboard, clipboard, and download results.

## Repository and release preparation

- [ ] Fresh setup/build works from documented commands and lockfiles.
- [ ] Build outputs, caches, local runtime data, and dependencies are ignored.
- [ ] Remove verified unused starter code/dependencies without breaking hosting.
- [ ] Required content, application, Go, and bundle-budget CI checks pass.
- [ ] Record unsupported toolchain/external-service scopes without marking them passed.
- [ ] Write contributor/setup/license/version documentation.
- [ ] Run V3 on port 4175 and prepare separate hosted comparison access.
- [ ] Verify original and V2 remain available and update comparison records.
- [ ] Deliver source/course ZIPs, before/after report, and final verification evidence.

## GitHub publication

- [ ] Resolve repository owner/name and inspect any existing remote repository.
- [ ] Review publishable files/history, license scopes, and credential references.
- [ ] User confirms the concrete destination and public publication.
- [ ] Create/connect repository and push without overwriting unrelated history.
- [ ] Verify remote files, required CI, release/version references, and repository URL.
- [ ] Record any separate public-hosting requirement and its destination decision.
