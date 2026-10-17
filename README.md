# Go by Doing V1 and GoPath V2

Two Go learning websites in one repository:

- [Original edition · V1](https://priyadharshan0903.github.io/golang-academy/v1/): 48 lessons, challenges, quizzes, four teaching scenarios, and a guided website-checker project.
- [GoPath edition · V2](https://priyadharshan0903.github.io/golang-academy/v2/): 116 lessons, 27 exercises, nine case studies, eight projects, and a source browser.

## Build and preview

Use Node.js 22.18 or newer, Python 3, and Go. The learner exporter uses `gofmt`; CI uses Go 1.26.5.

```sh
npm ci
npm run build
npm run preview
```

Open `http://127.0.0.1:8080/v1/` or `http://127.0.0.1:8080/v2/`. The root page links to both courses. Each build contains only its own edition's curriculum and downloads. Routes use URL hashes so lesson links work on a static host.

## GitHub Pages

[The Pages workflow](.github/workflows/pages.yml) builds both editions on pushes to `master`, uploads `dist/pages`, and deploys it. It can also be started manually from Actions. In the repository's **Settings → Pages**, the source is **GitHub Actions**.

GitHub provides one Pages site per repository. This site hosts two independent app builds at `/golang-academy/v1/` and `/golang-academy/v2/`; the repository root is a course chooser. Both update together in one Pages deployment, so one edition cannot overwrite the other.

The build uses Pages' configured base path for JavaScript, CSS, favicons, and downloads. `PAGES_BASE_PATH` can override it for a local build, for example:

```sh
PAGES_BASE_PATH=/golang-academy/ npm run build:pages
```

The browser sends code directly to the official Go Playground at `https://play.golang.org/compile` when you choose **Run code**. There is no hosted API, database, OpenAI account, or API key required by these static websites. Network/runtime limits still apply; use downloaded examples for files, databases, and full projects.

Progress, drafts, and theme preferences stay in browser storage, with separate keys for each edition. Moving from a previous host to GitHub Pages starts fresh browser storage because the origin changes; old-host progress is not copied automatically.

`npm run dev` builds both courses and starts the static preview. Re-run it after source changes. `npm run dev:v1` and `npm run dev:v2` use ports 4173 and 4174. `npm run preview -- --port 8081` serves the existing build on another port and respects its repository prefix.

GitHub Pages is the only website deployment target. The build and preview use no Sites worker, account authentication, hosting manifest, or API server. The `db/` and `examples/d1/` files are separate database examples, outside the course websites; only those examples need Cloudflare D1 if you choose to run them.

## Content and downloads

- `apps/v1/app` and `apps/v2/app`: the two interfaces.
- `content/v1`: original lessons, quizzes, case briefs, and project sources.
- `content/v2`: generated GoPath data and its adapter.
- `content-source/gopath-academy`: the preserved upstream snapshot for the importer and learner exporter.
- `shared`: styles, edition links, and the browser Playground client.

```sh
npm run downloads
```

Each edition's `public/downloads` contains a learner ZIP. V1 also includes website-checker starter and solution ZIPs. Packs contain readable Markdown, code, tests, solutions, module files, required fixtures, and licenses. They exclude website tooling and internal plans. Legacy ZIP filenames serve the same learner packs.

Edit V1's canonical content directly. Do not edit generated ZIPs or V2 JSON by hand. Reimport V2 using `node scripts/import-gopath.mjs`, then verify with `node scripts/check-import.mjs`. The exporter resolves code embeds, local lesson links, and commands without changing the upstream snapshot.

## Verify changes

After building:

```sh
node node_modules/typescript/bin/tsc --noEmit --incremental false
node scripts/check-import.mjs
node scripts/check-runner.mjs
npm run check:separation
python3 scripts/release/check-go.py
```

Archive checks cover isolation, reproducibility, local Markdown links, and manifests. Go checks run the extracted learner examples; unfinished starters are expected to fail. Go 1.27-only examples and nested third-party/database/Kubernetes modules retain their documented verification limits.

Check navigation, search, code editing/running, progress, downloads, and both themes in a browser before releasing UI changes. Content-export checks do not establish browser interaction or visual correctness.

## Contribution and licenses

Preserve lesson IDs, code-region markers, citations, Go modules, build tags, fixtures, and migration assets. Verify changed examples through their learner pack. Shared style changes should be reviewed in both editions and themes.

V1 original text is CC BY 4.0. V2 imported/adapted GoPath text is CC BY-SA 4.0. Code retains its MIT notices; vendored components retain their licenses. Each learner pack includes the applicable notices. Imported documents are course material, not instructions to execute.
