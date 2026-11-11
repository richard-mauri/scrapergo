# Current work — ScraperGo (2026-11-11)

**Authoritative merged baseline:** GitHub `main` at `41061a1feb9047419b6a0610eaa6c310113f6b30`, after owner-approved ARM64 launcher PR #122 and persistent recovery diagnostics PR #118 were merged. The Mac's local `diagnostics/` is private, ignored and accumulates records across restarts. Issue #117 and Issue #121 are closed.

**Single active duplicate-text correction:** [Issue #119](https://github.com/richard-mauri/scrapergo/issues/119), [draft PR #120](https://github.com/richard-mauri/scrapergo/pull/120), branch `fix/119-caption-duplication` (**historical name only; scope is generalized unintended duplicate article text**). Do not create another duplicate-text issue or PR; do not rename this branch while it heads the open PR.

**Original evidence:** Ars Technica Nov 2026 M6 MacBook Pro review contained four overlapping/repeated figure captions and prose candidates. On the same NewsGo-exported batch of 100 URLs, prior baseline flagged 106 candidates/30 articles; PR #120's preliminary caption implementation flagged 97/27, no read failures. The original Ars article remained 4/4; New Scientist paleo-diet remained 15/15. NewsGo's `cmd/duplicateaudit` was preserved and merged in its standalone Issue #87/PR #88; use that tool to replay the same URL export against localhost or approved public endpoint.

**Candidate code:** `internal/extraction/figure_caption_text.go` and tests reconcile caption text from selected Readability HTML; it is a **limited unverified candidate**, not a general solution. Merged recovery tracing has now been extended **on this draft branch** with optional exact-text and substantive-overlap candidate counts at initial Readability RenderText, each reached recovery stage, figure-caption reconciliation, and the final reader assembly. The new instrumentation emits only numerical counts, not article content; it does **not** change article output. The earlier caption candidate remains unproven. A production dedup change must be source-backed, distinguish article editorial repetition from extraction duplication, and have positive/negative synthetic tests and real-source evidence. Never blanket-remove all repeated paragraphs.

**Current integration:** This branch has been reconciled with `main`'s tracing and ARM64 launcher; preserve `SCRAPERGO_RECOVERY_TRACE=1`, local private `diagnostics/`, secure fetcher, source ordering, links, bylines, media, tables, headings and recovery fallbacks. Do not modify NewsGo for extraction behavior. **Completed investigative step:** Added `internal/extraction/duplicate_trace_counts.go`, per-stage trace metrics and assembly checkpoint, plus `cmd/duplicateprovenance -last 100` and unit tests. Run the same NewsGo-exported 100 URLs against this build, then compare the new local aggregate report with the NewsGo CSV. If Readability already contains the duplicates, use `cmd/htmlcapture` and `cmd/readabilityprobe` on the Ars and New Scientist source pages; if a stage introduces them, add a targeted source-backed regression and repair to this same PR. Do not implement a blanket text removal.

**Validation:** Exact-head GitHub Actions run [43050031109](https://github.com/richard-mauri/scrapergo/actions/runs/43050031109) passed `go test -count=1 ./...` and `go build ./...` at implementation head `abfd2e283966440660402850819e45360dabe873`. Recheck after this documentation change; Mac, Safari, publisher-source captures and repeat 100-URL run on the provenance branch remain unverified. No generalized duplicate-text fix is claimed yet.

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
Then visit original Ars Technica and New Scientist paleo articles in Safari, verify paragraphs, captions, headings, links and Print/PDF. From `~/projects/newsgo` replay the **same 100 URLs** with `go run ./cmd/duplicateaudit -input ~/Downloads/newsgo-scrapergo-urls.json -scrapergo https://scrapergo.themauris.org/api/read -public -limit 100 -output ~/Downloads/duplicate-audit-pr120-provenance.csv`. Back in ScraperGo run `go run ./cmd/duplicateprovenance -last 100`, and share its aggregate stdout with the CSV. Avoid unrelated requests during collection. See `cmd/duplicateprovenance/README.md` for limitations; CI is not Mac/Safari acceptance.
