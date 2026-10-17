# Public repository security review

Reviewed 2026-10-17. Compatible dependency fixes are applied to `package.json`
and `package-lock.json`. No live credential was detected in the scanned
publication inputs. This review is not a guarantee that the repository or an
external account is free of all security issues.

## Scope and secret findings

- Scanned 838 tracked/unignored publication files, their 1,174 ZIP entries,
  both editions' generated learner packs, built production artifacts,
  repository remote configuration, and all reachable Git history (one commit
  at the final check). Gitleaks 8.30.1 was verified against its release checksum.
- Environment files and private account directories were not opened. Ignored
  local caches, dependency folders, and general account state are outside this
  audit. No account login history, billing activity, or API key inventory was
  available for inspection.
- No OpenAI API keys/session tokens, GitHub tokens, cloud access keys, or private
  keys were detected. Generic credential matches were local teaching database
  examples and a test fixture. Gitleaks' eight original generic-key matches were
  Go module `h1:` integrity hashes embedded in curriculum JSON.
- `.gitleaks.toml` permits only that specific checksum syntax for the generic-key
  rule. It does not exclude the curriculum file or disable provider key rules.
  A synthetic API key still causes Gitleaks to fail with this configuration.
- Artifact scans flagged each build's framework prerender token in the two
  server-only `vinext-server.json` manifests. These are internal server tokens,
  not account credentials. Their values are absent from the publication file
  list and client assets. They remain outside learner packs and the public Git
  tree; they were not added to the Gitleaks allowlist. Publish only client assets
  to a static host, never the complete server output directory.
- OpenAI Site project IDs, version IDs, and published Site addresses remain in
  hosting configuration and release documentation. These identify deployments;
  they are not bearer tokens or API credentials. One machine-specific absolute
  path was removed from the portable build script.
- The learning app's Go runner sends source to the fixed Go Playground compile
  endpoint. No OpenAI API integration or account-billing credential was found
  in that request path. Development preview authentication/connector helpers
  are limited to development server hooks.

## Dependency fixes and remaining findings

Counts below are npm's affected dependency entries, including propagated
findings, rather than counts of distinct vulnerabilities.

| Audit scope | Before | After |
| --- | --- | --- |
| Production dependency classification (`--omit=dev`) | 9: 1 critical, 3 high, 5 moderate | 0 |
| Complete dependency tree | 66: 1 critical, 46 high, 17 moderate, 2 low | 12: 8 high, 4 moderate |

Updated Next.js and its ESLint config from 16.3.4 to 16.3.8; React, React DOM,
and React Server Components from 19.2.6 to 19.2.8; and Vite from 8.0.13 to
8.0.16. Compatible transitive fixes include Sharp, Undici, ws, image-size,
source-map-js, Babel, js-yaml, brace-expansion, fast-uri, and fflate.
Scoped overrides repair pinned transitive packages without forcing a framework
major upgrade. Lockfile integrity and a clean `npm ci --ignore-scripts` passed.

The remaining affected packages are in the tooling dependency graph:

- **braces 3.0.3, high:** deeply nested patterns can exhaust the stack.
  The registry has no patched release. It propagates through micromatch,
  fast-glob, ESLint's Next plugin, and Vinext's CommonJS import tooling.
  [Advisory GHSA-vfj7-8cjw-p6xm](https://github.com/advisories/GHSA-vfj7-8cjw-p6xm).
- **esbuild 0.18.20, moderate:** the development server permits cross-origin
  response access. This copy is pinned by Drizzle Kit's deprecated
  `@esbuild-kit` loader. The compatible update pass retains it; replacing that
  loader or forcing a different esbuild minor needs a separate compatibility
  change. The newer esbuild copies used by Wrangler/Vite/tsx are patched.
  [Advisory GHSA-67mh-4wv8-2f99](https://github.com/advisories/GHSA-67mh-4wv8-2f99).

No outstanding package markers for braces or esbuild were found in the built
server JavaScript. This is supporting evidence, not a complete exploitability
analysis. Development dependencies can affect build security even when the
production-only npm audit is clean. The full GitHub dependency check therefore
remains intentionally failing while the high tooling advisory is unresolved.
No forced downgrades or breaking dependency fixes were applied.

## Safeguards and verification

- Recursive ignores cover dependency/build/tool state, environment files,
  private key files, credential files, and Python/TypeScript caches.
- `npm run check:secrets` scans the Git publication file list and learner ZIPs,
  reports locations/rules without values, and rejects protected files by name
  without reading them. It supplements Gitleaks; it is not a complete detector
  for arbitrary credential formats.
- `.github/workflows/security.yml` adds read-only push/PR checks for history,
  current files, archives, and production/full dependency audits. Actions are
  pinned by commit; Gitleaks is pinned by version and SHA-256. Checkout does not
  retain credentials. These checks need no OpenAI key or repository secret.
- Both isolated production builds passed with the patched clean installation.
  TypeScript, all 16 runner checks, edition isolation, deterministic archives,
  local Markdown links, ZIP integrity, and file manifests passed. Local HTTP
  checks passed for both roots, matching learner/legacy ZIPs, edition download
  isolation, and malformed runner requests. Workflow YAML parsed successfully.
- Current-file and history Gitleaks scans passed; production artifact findings
  were separately reviewed as described above. Six generated edition ZIPs
  (1,530 entries, including aliases and project downloads) passed the provider
  key/private-key guard. The lightweight guard and
  Gitleaks both detected synthetic provider keys during positive-control checks.

The existing workspace `node_modules` directory could not be replaced: macOS
returned `EPERM`, including after an elevated retry. Verification used a fresh
installation in a temporary directory, selected explicitly through the edition
builder's `dependencyDirectory` option. Run `npm ci` in a normal terminal or a
fresh checkout before using the updated dependencies for routine local work.
The current hosted releases have not been redeployed by this audit.

Raw redacted audit reports and build logs are kept in ignored
`work/security-audit/`, outside publication inputs.

## Publication tracker

- [x] Scan current publication files and learner archives for credentials.
- [x] Scan all currently reachable Git history.
- [x] Triage teaching fixtures and narrowly allow Go checksum false positives.
- [x] Patch compatible dependencies and verify both edition builds.
- [x] Add reusable secret checks and a read-only security workflow.
- [ ] Replace the blocked local dependency installation with a fresh `npm ci`.
- [ ] Resolve the remaining high build-tool advisory when an upstream fix or
      reviewed replacement is available; replace Drizzle's legacy loader.
- [ ] Run the workflow on GitHub and configure repository secret scanning/push
      protection where available. No GitHub repository settings were changed.
- [ ] Before exposing a hosted Go runner publicly, add deployment-level request
      limits. Its size/origin/timeout checks do not authenticate or rate-limit
      arbitrary clients. GitHub Pages will require a separate runner backend or
      an external Playground link because it cannot serve this API route.
