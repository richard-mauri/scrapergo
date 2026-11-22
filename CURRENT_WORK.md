# Current work — ScraperGo

**Merged main verified at session start:** `e22b2e7427a6884074d10e72b24b47fe544aded8`. **Current scope:** [Issue #147](https://github.com/richard-mauri/scrapergo/issues/147), branch `fix/147-newsgo-thumbnail-auth` based on merged `main`. Related but independent NewsGo [Issue #114](https://github.com/richard-mauri/newsgo/issues/114) / draft [PR #115](https://github.com/richard-mauri/newsgo/pull/115). Previous `auth/145-verified-suite-access` branch was absent and is not the active work branch.

## Cause and change

NewsGo Safari demonstrated 190 thumbnail failures out of 190 attempts, with ScraperGo `GET /api/thumbnail` returning HTTP 401 to a naked browser `<img>`. Article classification and RSS metadata-only publisher images still worked. ScraperGo now enforces Homepage per-request API auth, and only had a dedicated machine exception for NewsGo `POST /api/classify`.

This independent ScraperGo draft extends `internal/app/auth_gate.go` with an **independent** private machine credential `SUITE_NEWSGO_SCRAPERGO_THUMBNAIL_TOKEN`. A valid lowercase-hex `Authorization: Service` token can call **only `GET /api/thumbnail`**, with no Origin header and no reused classify credential. All other machine requests fail through the existing verified Homepage bearer guard; ordinary personal/Guest media behavior is unchanged. Existing thumbnail retrieval, SSRF, resizing, caching, article extraction and publisher responses are untouched.

## Regressions and integration

Added `internal/app/thumbnail_service_auth_test.go`: positive GET/thumbnail and established POST/classify, negatives across methods, identity, read, PDF, classify, image, diagnostics, activity, unknown routes, wrong/missing/reused tokens and Origin restrictions. Documentation reviewed: Homepage merged `SUITE_DEVELOPMENT_POLICY.md` v1.1.0; ScraperGo AGENTS, README, ARCHITECTURE, DEVELOPMENT, CURRENT_WORK. Updated README, ARCHITECTURE, DEVELOPMENT, CURRENT_WORK; AGENTS and Homepage policy unchanged.

**Required matching NewsGo PR #115** makes authenticated user-scoped requests to a same-origin proxy, which forwards the distinct thumbnail service credential server-to-server. No credential values have been generated or installed. Owner Mac uncommitted state, private configuration, tunnel, running process, build counter and live publisher data are not accessible. No production service, tunnel, private token, account, data or Safari state modified. GitHub Go CI run #788 **passed** on code head `1b9c2fe53638c10d96fd55fd3a3a5b078aab5012`, including Go tests/build, JS checks and offline diagnostics. Subsequent comment/documentation commits require their own exact-head confirmation. Local Mac `./devbuild.sh`, coordinated Safari personal/Guest and live image tests remain unverified. Promotion requires explicit separate owner authorization.

## Owner Mac handoff (candidate only; no restart permission)

```bash
cd ~/projects/scrapergo
git status --short --branch
# Stop if any uncommitted changes are shown.
git fetch origin
git checkout fix/147-newsgo-thumbnail-auth
git pull --ff-only origin fix/147-newsgo-thumbnail-auth
./devbuild.sh
./dist/scrapergo-darwin-arm64 --version
# Do not run ./scrapergo.sh without distinct production restart authorization.
```

No merge, Issue closure, remote branch deletion, deployment, restart, version/build-counter change, tag or release authorized.
