# V1 and V2 website separation checklist

This records the earlier Sites separation and private release. Current public hosting uses two static builds on GitHub Pages; see [the deployment instructions](../../README.md#github-pages).

Status: Both websites are deployed at their existing origins with separate, verified learner downloads. Implementation, archive, build, Go, runner, and HTTP release checks are complete. Five real-browser verification items remain pending because macOS blocks Chromium launch.

Plan: [V1 and V2 website separation plan](V1-V2-SPLIT-PLAN.md).

Check an item only after its evidence is recorded. Keep supporting reports under a dedicated separation-review directory when implementation begins. Record blockers and external-service/toolchain limits without marking those checks passed.

## Planning

- [x] Inspect V1/V2 app entry points, content sources, state keys, and hosting configuration.
- [x] Inspect current download generators and the 619-file V2 archive.
- [x] Identify learner dependencies that a blanket folder exclusion would break.
- [x] Define two app builds, edition-specific learner downloads, migration phases, and release criteria.
- [x] Save a separate plan and checklist alongside the existing V3 documents.

## Phase 1 Inventory and build proof

- [x] S01 Verify the actual V1 and V2 hosted URLs, Site identities, audiences, and source locations.
- [x] S02 Record both curriculum inventories, IDs, ordering, routes, licenses, and generated files.
- [x] S03 Map app imports, static assets, renderer dependencies, and runner endpoints.
- [x] S04 Map progress, drafts, project checkpoints, theme, and sidebar storage keys.
- [x] S05 Record Go modules, fixtures, embedded assets, migrations, code regions, and relative commands.
- [x] S06 Identify website-only files versus lesson-required deployment examples.
- [x] S07 Prove independent V1 and V2 builds and supported publication roots without changing live sites.

Evidence: `docs/separation-review/BOUNDARIES.json`, `STATE-AND-CONTENT.json`, and both independent production build results in `work/separation/`. Existing Site identities and owner-private audiences verified through Sites.

## Phase 2 App separation

- [x] S08 Establish separate V1 and V2 app roots and build commands.
- [x] S09 Give each app its own entry point, metadata, public assets, download outputs, and hosting manifest.
- [x] S10 Preserve V1 curriculum, quizzes, cases, and guided website-checker.
- [x] S11 Preserve V2 lessons, exercises, cases, projects, embeds, source browser, and provenance.
- [x] S12 Remove cross-edition curriculum imports and unrelated public assets from each build.
- [ ] S13 Preserve state keys and lesson IDs; verify existing saved state still loads.
- [x] S14 Preserve origins or implement validated state export/import with conflict handling.
- [ ] S15 Preserve existing routes, heading links, and browser navigation.
- [x] S16 Configure working edition links and runner endpoints for both apps.
- [x] S17 Verify V1 and V2 can each build without the other's curriculum.

Evidence: independent standalone build roots and `STATE-AND-CONTENT.json`. Content and storage-key bytes are preserved. S13 and S15 remain unchecked until stored-state and browser navigation can be exercised in a real browser.

## Phase 3 Learner downloads

- [x] S18 Create separate content exporters and explicit file allowlists for V1 and V2.
- [x] S19 Export standard Markdown with ordered indexes, readable code embeds, and working relative links.
- [x] S20 Preserve V1 explanations, expected outputs, practice, hints, solutions, quizzes, cases, and projects.
- [x] S21 Preserve V2 lessons, case citations, exercise starters/tests/solutions, and project source.
- [x] S22 Include required Go module files, fixtures, embedded assets, migrations, and learner configuration.
- [x] S23 Preserve Go module layouts and exercise build-tag behavior.
- [x] S24 Rewrite and verify commands affected by exported paths, including the word-frequency checkpoint.
- [x] S25 Exclude website UI/build/publishing files, unrelated CI, internal plans, caches, and runtime state.
- [x] S26 Preserve appropriate license notices and attribution in each pack.
- [x] S27 Generate `go-by-doing-v1-content.zip` and `gopath-v2-content.zip` from fresh staging directories.
- [x] S28 Record edition, release/source revision, file inventory, file hashes, and archive checksum.
- [x] S29 Confirm identical inputs produce identical archive checksums.
- [x] S30 Verify V1 starter and solution downloads remain available and correct.
- [x] S31 Unzip both packs cleanly and read lessons without Node.js, React, Vinext, or Hugo.
- [x] S32 Run supported examples, project tests, and reference solutions from the extracted packs.
- [x] S33 Record intentionally failing starter tests and unsupported toolchain/service checks separately.

Evidence: `ARCHIVE-CHECKS.json`, `GO-CHECKS.json`, and each ZIP manifest. All 52 scoped Go checks pass. Toolchain/service exclusions are recorded in `GO-CHECKS.json`.

## Phase 4 Download experience

- [x] S34 Update each header, reader, exercise, project, source-browser, and Resources download link.
- [x] S35 Use clear learner-facing labels such as “Download learning materials” with accurate contents.
- [x] S36 Ensure each site serves only its own edition's learner pack.
- [x] S37 Preserve old archive URLs through compatible files or redirects.
- [x] S38 Verify download MIME types, filenames, ZIP integrity, and matching release metadata.

Evidence: `HTTP-CHECKS.json`; every download URL found in both app sources returned a valid ZIP with `application/zip`. Legacy URLs serve identical learner-pack bytes.

## Phase 5 Verification and release

- [x] S39 Pass content checks, TypeScript checks, and production builds for both apps.
- [ ] S40 Verify search, completion/resume, drafts, projects, source browsing, and runner behavior.
- [ ] S41 Check representative desktop/mobile flows in light and dark themes and with keyboard navigation.
- [ ] S42 Verify direct links, legacy links, edition switching, and every visible download.
- [x] S43 Confirm app assets and public directories exclude the other edition's content/downloads.
- [x] S44 Confirm final hostnames, Site identities, current audiences, and origin/state handling.
- [x] S45 Record independent release and rollback references, including matching content ZIPs.
- [x] S46 Deploy V1 to its confirmed destination and verify its website and learner ZIP.
- [x] S47 Deploy V2 to its confirmed destination and verify its website and learner ZIP.
- [x] S48 Verify both websites remain available after each release.
- [x] S49 Record deployed URLs, source/content revisions, archive checksums, and remaining limitations.
- [x] S50 Update setup, content export, release, and contributor documentation for both app roots.

Evidence: `docs/separation-review/RELEASES.json`, hosted ZIP checksum matches, `HTTP-CHECKS.json`, and `VERIFICATION.md`. Browser checks remain explicitly pending; they are not inferred from HTTP or build results.

## Release tracker

| Item | V1 | V2 |
| --- | --- | --- |
| Website URL and Site identity | https://go-by-doing.workday-code-8517.chatgpt.site · appgprj_6ad1ec4b52d0819182e57f2c75418cb4 | https://go-by-doing-gopath-v2.workday-code-8517.chatgpt.site · appgprj_6ad25f3fe054819183e766eb2c2433fa |
| Audience | Owner-private, preserved | Owner-private, preserved |
| Source/content revision | ec2354839086 | 573c74b5daea |
| Independent build | Passed | Passed |
| Learner ZIP and checksum | 728fb2afc6e4da9fd9d8f1751576ab5d157c6e9047fc0455547ef3669266d44c | 4b332e6d889ce5cb5784ec51b28e06739d1965a32bbbeb5da4ef068f16f86a62 |
| Content and Go verification | Passed; scope recorded | Passed; scope recorded |
| Browser verification | Blocked: macOS Chromium launch permissions | Blocked: macOS Chromium launch permissions |
| State and link compatibility | Keys/IDs/origin preserved; HTTP links passed; browser checks pending | Keys/IDs/origin preserved; HTTP links passed; browser checks pending |
| Deployment verification | Succeeded; hosted ZIP checksum matched | Succeeded; hosted ZIP checksum matched |
| Rollback reference | appgprj_6ad1ec4b52d0819182e57f2c75418cb4~appgver_21f8591ea54081919cc4c07b7724a3ee | appgprj_6ad25f3fe054819183e766eb2c2433fa~appgver_52c0946bef988191bd896f5812bcc81c |

## Work log

| Date | Item | Result or evidence | Next action |
| --- | --- | --- | --- |
| 2026-10-17 | Planning | Inspected app/data boundaries, download generators, ZIP contents, module files, relative commands, and existing review plans. | Implement approved separation. |
| 2026-10-17 | Implementation | Extracted app/content roots, built each independently, generated deterministic learner packs, and passed archive/link/Go/runner/HTTP checks. Canonical content and state-key strings match previous hosted source. | Complete releases; retain browser verification as a documented gap. |
| 2026-10-17 | Release | V1 and V2 deployed at existing URLs. Both hosted ZIPs and legacy downloads match verified local archives; V1 rechecked after V2 publication. Rollback versions recorded. | Complete S13, S15, S40, S41, and S42 when real-browser access is available. |
