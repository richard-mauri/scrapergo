# Current work — ScraperGo

**Merged main verified:** `e22b2e7427a6884074d10e72b24b47fe544aded8`. **Active scope:** [Issue #147](https://github.com/richard-mauri/scrapergo/issues/147), draft [PR #148](https://github.com/richard-mauri/scrapergo/pull/148), branch `fix/147-newsgo-thumbnail-auth` based on `main`. Matching NewsGo [Issue #114](https://github.com/richard-mauri/newsgo/issues/114), draft [PR #115](https://github.com/richard-mauri/newsgo/pull/115), branch `fix/114-authenticated-thumbnails`. NewsGo merged main verified at `cc86480caafbf3299b1064c380e0d7e5bd17ff3c`. Historical auth PR #146 / branch `auth/145-verified-suite-access` is no longer active.

## Purpose and current contract

Owner Safari observed all 190 NewsGo thumbnails fail because an ordinary browser `<img>` request to ScraperGo `GET /api/thumbnail` returned HTTP 401. ScraperGo correctly requires a Homepage personal/Guest bearer on its protected APIs; classification was separately authorized with a private NewsGo machine credential.

Owner recovered and confirmed the **existing** private `SUITE_NEWSGO_SCRAPERGO_CLASSIFY_TOKEN` in `/Users/richardmauri/projects/homepage/private/service-auth.env` and directed that it be reused for all specifically authorized NewsGo→ScraperGo machine APIs. This scoped draft extends the existing machine credential's authorization to exactly **`POST /api/classify` and `GET /api/thumbnail`**. Both operations require the strict 64-lowercase-hex `Authorization: Service` credential, constant-time comparison and no Origin header. All other paths and methods (reader, PDF, article-image, metadata, identity, diagnostics, activity, unknown routes) still require independently verified Homepage personal/Guest access. No `SUITE_NEWSGO_SCRAPERGO_THUMBNAIL_TOKEN` variable or additional token is needed.

NewsGo PR #115 is the matching protected same-origin proxy and temporary image-blob client. Existing ScraperGo thumbnail handling, publisher SSRF validation, caching, extraction, classification, image/PDF behavior, browser authentication, per-user isolation and Safari extension are unchanged.

## Verification and constraints

Added Go regression tests ensuring the **same** machine token can use both permitted operations while missing, malformed, wrong or Origin-bearing machine requests and unrelated routes/methods are denied. Homepage `SUITE_DEVELOPMENT_POLICY.md` v1.1.0 read directly from merged `main`. Reviewed ScraperGo AGENTS/README/ARCHITECTURE/DEVELOPMENT/CURRENT_WORK; updated README, ARCHITECTURE, DEVELOPMENT and this handoff; AGENTS and Homepage policy unchanged. **Verify latest exact-head CI before handoff**; previous CI #790 was on an earlier candidate and does not verify the current change.

Owner Mac's uncommitted/unpushed work, private token value, local run state, publisher data, build counters, tunnel and browser state remain uninspected. No token creation/configuration, persistent data, processes, tunnel, owner-local settings, tags, releases or app version/build counters changed. Mac `./devbuild.sh`, Safari personal/Guest real images and coordinated NewsGo+ScraperGo runtime acceptance are unverified.

## Owner Mac handoff (after CI passes)

```bash
cd ~/projects/scrapergo
git status --short --branch
# Stop if any uncommitted changes appear.
git fetch origin
git checkout fix/147-newsgo-thumbnail-auth
git pull --ff-only origin fix/147-newsgo-thumbnail-auth
./devbuild.sh
./dist/scrapergo-darwin-arm64 --version
# Run ./scrapergo.sh only after separate production restart authorization.
```

The owner uses a **2026 M6 Apple Silicon MacBook Pro**; use the ARM64 executable. `./devbuild.sh` runs Go tests, benchmark smoke and both macOS builds. The version command is optional and need not precede launch.

**Promotion authority:** Draft only; no FINISH, merge, issue closure, remote branch cleanup, deployment/restart, token generation/change, tag, release or semantic version bump.
