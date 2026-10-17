# Go by Doing V1 and GoPath V2

Two independently deployable Go learning websites maintained in one source workspace. V1 contains 48 original lessons, four teaching scenarios, quizzes, and a guided website-checker. V2 contains 116 GoPath lessons, 27 exercises, nine case studies, eight projects, and a source browser.

## Source boundaries

- `apps/v1/app`: original app, with its own metadata and hosting identity.
- `apps/v2/app`: GoPath app, with its own metadata and hosting identity.
- `content/v1`: original curriculum, project guide, source viewer data, and project source.
- `content/v2`: the generated GoPath data and its adapter.
- `content-source/gopath-academy`: immutable, hashed upstream V2 source snapshot used by the importer and learner exporter.
- `shared`: styles, edition URLs, and the Go Playground runner. Each app keeps its own state namespaces and lesson IDs.

The root app remains a convenient V2 entry point. Separate release builds contain only the chosen app and its content; they do not depend on the other edition.

## Install and build

Use Node.js 22.18 or newer, Python 3, and Go for formatting/export checks. The current verification toolchain is Go 1.26.5. Install dependencies once from this workspace:

```sh
npm ci
npm run build:v1
npm run build:v2
```

Each command generates its learner pack and builds in a fresh temporary standalone project. The build location is recorded in `work/separation/v1-build.json` or `v2-build.json`. This avoids filesystem replacement restrictions in some desktop environments and proves that each app can be published on its own.

```sh
npm run start:v1
npm run start:v2
```

Default ports are 4173 and 4174. If another review instance already uses those ports, supply alternatives:

```sh
npm run start:v1 -- 4183
npm run start:v2 -- 4184
```

Use `npm run dev:v1` or `npm run dev:v2` for an isolated development preview. Stop and recreate an isolated preview after changing canonical workspace sources; its source tree is a prepared snapshot.

## Learner downloads

```sh
npm run downloads
npm run check:separation
```

Outputs are generated under each edition's `public/downloads` directory:

- V1: `go-by-doing-v1-content.zip`, plus website-checker starter and solution ZIPs.
- V2: `gopath-v2-content.zip`.

Legacy archive URLs continue serving the corresponding learner pack. Exports contain readable Markdown, code, tests, solutions, module files, required fixtures, and licenses. They exclude website UI/build files, Hugo configuration/templates, CI, and internal plans. The V2 Kubernetes/server teaching examples remain included. See the generated README, index, and checksum manifest in each ZIP.

Do not edit generated ZIPs or V2 JSON by hand. Edit V1's canonical content directly. Reimport V2 with `node scripts/import-gopath.mjs` and verify with `node scripts/check-import.mjs`; this preserves the upstream snapshot hashes. Run the exporters after content changes. They resolve all code embeds and rewrite local lesson links and export-specific commands.

## Verification

```sh
node node_modules/typescript/bin/tsc --noEmit --incremental false
node scripts/check-import.mjs
node scripts/check-runner.mjs
npm run check:separation
python3 scripts/release/check-go.py
```

The separation check verifies independent build inputs, reproducible archive checksums, Markdown links, and ZIP file manifests. Run `check-go.py` after separation checks; it tests the actual extracted archives. Intentionally failing starter tests are checked separately. Go 1.27-only, nested third-party, database/Redis, and Kubernetes execution retain their documented verification limits.

For local HTTP checks, run both built apps on ports 4183/4184, then `python3 scripts/release/check-http.py`. Real browser verification is recorded separately; HTTP checks do not establish visual layout, keyboard interaction, or stored-state behavior.

## Publication

Existing Site identities are retained in `apps/v1/.openai/hosting.json` and `apps/v2/.openai/hosting.json`. Current origins preserve access to each edition's browser storage. The workspace's current Sites access policy supports private review, not public publication.

Publication checkouts are generated snapshots, separate from the maintained source workspace. Open each existing Site with the Sites source helper before preparation. After successful builds, `node scripts/release/prepare-publication.mjs v1` or `v2` synchronizes that edition into `work/publish/<edition>` with its matching build output. The source helper then commits/pushes and packages the exact source and artifact; deployment uses the saved version for that Site.

Releases, verification evidence, limitations, and rollback references are tracked in [the separation checklist](docs/review-plans/V1-V2-SPLIT-CHECKLIST.md) and `docs/separation-review/`.

## Content and contribution

Keep lessons, IDs, code-region markers, source citations, nested Go modules, build tags, fixtures, and migration assets intact. Verify a changed example through its exported learner pack, and keep local commands consistent with the exported layout. UI changes shared through `shared/styles/globals.css` need review in both editions and themes.

V1 original text is CC BY 4.0. V2 imported/adapted GoPath text is CC BY-SA 4.0. Code retains its MIT notices; vendored components retain their own licenses. Each learner pack carries its own applicable notices. The V2 source importer and export scripts treat imported documents as course material, not instructions to execute.
