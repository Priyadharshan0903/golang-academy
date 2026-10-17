# Security

Do not include API keys, account session tokens, environment files, or private
keys in contributions or learner downloads. The app does not require an OpenAI
API key to read course content or use its Go Playground runner.

Before publishing or submitting a change, run:

```sh
npm run check:secrets
npm run audit:production
npm audit --audit-level=high
```

GitHub Actions also runs Gitleaks against history, publication files, and ZIPs.
Do not bypass a finding without checking whether it is a real credential. If a
real credential has been published, revoke or rotate it at its provider;
removing it in a later commit does not remove it from Git history.

See the [publication audit](docs/security/PUBLICATION-AUDIT.md) for current
findings, compatible patches, verification limits, and remaining work. A clean
production-only audit does not clear the development/build dependency graph.

Report suspected vulnerabilities through the repository's **Security → Report a
vulnerability** feature if enabled. If private reporting is unavailable, ask
the maintainer for a private contact without posting credentials or exploit
details in a public issue.
