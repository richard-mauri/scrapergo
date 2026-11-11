# Current work — ScraperGo (2026-11-11)

**Authoritative merged baseline:** GitHub `main` at `41061a1feb9047419b6a0610eaa6c310113f6b30`, after owner-approved ARM64 launcher PR #122 and persistent recovery diagnostics PR #118 were merged. The Mac's local `diagnostics/` is private, ignored and accumulates records across restarts. Issue #117 and Issue #121 are closed.

**Single active duplicate-text correction:** [Issue #119](https://github.com/richard-mauri/scrapergo/issues/119), [draft PR #120](https://github.com/richard-mauri/scrapergo/pull/120), branch `fix/119-caption-duplication` (**historical name only; scope is generalized unintended duplicate article text**). Do not create another duplicate-text issue or PR; do not rename this branch while it heads the open PR.

**Original evidence:** Ars Technica Nov 2026 M6 MacBook Pro review contained four overlapping/repeated figure captions and prose candidates. On the same NewsGo-exported batch of 100 URLs, prior baseline flagged 106 candidates/30 articles; PR #120's preliminary caption implementation flagged 97/27, no read failures. The original Ars article remained 4/4; New Scientist paleo-diet remained 15/15. NewsGo's `cmd/duplicateaudit` was preserved and merged in its standalone Issue #87/PR #88; use that tool to replay the same URL export against localhost or approved public endpoint.

**Candidate code:** `internal/extraction/figure_caption_text.go` and tests reconcile caption text from selected Readability HTML; it is a **limited unverified candidate**, not a general solution. Merged diagnostics trace records stage modifications but not which stage introduces duplicate text. A production dedup change must be source-backed, distinguish article editorial repetition from extraction duplication, and have positive/negative synthetic tests and real-source evidence. Never blanket-remove all repeated paragraphs.

**Current integration:** This branch has been reconciled with `main`'s tracing and ARM64 launcher; preserve `SCRAPERGO_RECOVERY_TRACE=1`, local private `diagnostics/`, secure fetcher, source ordering, links, bylines, media, tables, headings and recovery fallbacks. Do not modify NewsGo for extraction behavior. Proposed next step: add a privacy-minimized per-stage duplicate-candidate count to diagnostic trace to determine whether repeats originate in Readability text or recovery transformations, then use targeted source HTML capture/readabilityprobe for Ars and New Scientist before changing runtime removal rules.

**Deferred:** Issue #116 — performance/shared-DOM work. No PR #120 merge, release, semantic-version bump, deploy, branch deletion or app restart without owner acceptance.

**Mac commands for PR #120 test:**
```sh
cd ~/projects/scrapergo
git status --short --branch
git fetch origin
git checkout fix/119-caption-duplication
git pull --ff-only origin fix/119-caption-duplication
go test -count=1 ./...
./devbuild.sh
./scrapergo.sh
```
Then visit original Ars Technica and New Scientist paleo articles in Safari, verify paragraphs, image captions, headings, links and Print/PDF; replay 100 URLs with NewsGo's `cmd/duplicateaudit`, comparing by URL and separating failed reads. No local build, Safari acceptance or publisher-source probe is proven merely by CI.
