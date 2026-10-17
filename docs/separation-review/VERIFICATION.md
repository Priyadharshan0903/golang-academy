# V1 and V2 separation verification

Both editions were deployed on 17 October 2026 to their existing owner-private Site origins. The release IDs, source revisions, hosted archive checksums, and previous saved versions are in [RELEASES.json](RELEASES.json).

## Passed checks

| Area | Evidence |
| --- | --- |
| Independent builds | V1 and V2 each compiled from a fresh standalone root containing only its own app and content. Their hosting identities remained distinct. |
| Content preservation | [STATE-AND-CONTENT.json](STATE-AND-CONTENT.json): canonical curriculum bytes and storage-key strings match the previous hosted source revisions. |
| V2 import integrity | [IMPORT-CHECKS.json](IMPORT-CHECKS.json): 116 lessons, 27 exercises, nine cases, eight projects, 425 embeds, 445 source files; no broken source links or modified upstream snapshot files. |
| Learner packs | [ARCHIVE-CHECKS.json](ARCHIVE-CHECKS.json): deterministic ZIP checksums, readable Markdown, valid local links, complete file manifests, and separate edition inputs. |
| Extracted Go materials | [GO-CHECKS.json](GO-CHECKS.json): all 52 scoped checks passed using Go 1.26.5. This includes 48 V1 examples, the V1 project solution with the race detector, expected starter-test failures, V2 root-module packages/server, and V2 exercise reference solutions. |
| Runner behavior | [RUNNER-CHECKS.json](RUNNER-CHECKS.json): all 16 checks passed for request limits, origin validation, upstream handling, compiler errors, and failure paths. |
| Local HTTP | [HTTP-CHECKS.json](HTTP-CHECKS.json): both homepages and learner ZIPs returned 200; legacy ZIP bytes matched; the other edition's archive returned 404; malformed runner requests returned 400. All download links extracted from both app source trees returned valid ZIPs with `application/zip`. |
| Hosted release | Both homepages returned 200 with the expected download link. Both live learner ZIPs matched the verified local SHA-256 checksum. Legacy archive URLs returned the same learner-pack bytes. V1 was rechecked after V2 publication. |

V1's learner ZIP has 170 files including its manifest. V2's has 587 files including its manifest. Website UI, Hugo publishing files, internal plans, and website CI are excluded. Required Go module files, fixtures, SQL/assets, project files, and the Kubernetes/server teaching examples are retained.

## Browser verification still required

macOS denied Chromium launch, including an attempt using the available escalation mechanism. No passing browser interaction or screenshot review is claimed.

The [tracking checklist](../review-plans/V1-V2-SPLIT-CHECKLIST.md) leaves S13, S15, S40, S41, and S42 unchecked. Once browser access is available, exercise stored completion/resume and drafts, direct/hash navigation and history, search and project/source interactions, code copy/run, keyboard navigation, mobile layouts, both themes, edition switching, and downloads. Existing origins and state keys were preserved, but this does not substitute for testing real stored state.

Go 1.27-only code and nested modules requiring third-party dependencies, PostgreSQL/Redis, or Kubernetes remain outside the local Go verification scope. Starter tests intentionally fail until learner implementation; that behavior was verified separately.

## Reproduce the checks

Use Node.js 22.18 or newer, Python 3, and Go:

```sh
npm run build:v1
npm run build:v2
npm run check:separation
node scripts/check-import.mjs
node scripts/check-runner.mjs
python3 scripts/release/check-go.py
```

Start the built editions on unused ports for HTTP review. The existing review servers on 4173/4174 were preserved during this work; temporary verification servers used 4183/4184.

```sh
npm run start:v1 -- 4183
npm run start:v2 -- 4184
python3 scripts/release/check-http.py
```

Stop temporary servers after verification. The retained `work/separation` build reports point to local temporary build directories; regenerate the build if those directories have been cleaned.

## Rollback

Each release record contains a `rollbackVersionId` for that edition's previously saved version. Read the Site's current audience, then redeploy that exact saved version through the appropriate Sites deployment operation. Its saved build includes the corresponding older download artifacts. Roll back only the affected Site, then verify its homepage/download and confirm the other Site remains available. Restoring source separately uses `previousSourceCommit`; do not force-push unrelated source history.
