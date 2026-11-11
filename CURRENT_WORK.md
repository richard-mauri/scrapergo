# CURRENT WORK — ScraperGo handoff (November 11, 2026)

## Authoritative state and approval boundary

GitHub `main` is the authoritative **merged** baseline at `41061a1feb9047419b6a0610eaa6c310113f6b30`. PR #122 (native ARM64 launcher) and PR #118 (persistent opt-in recovery diagnostics) were owner-authorized and merged; Issues #121 and #117 were closed. The Mac launcher now runs `dist/scrapergo-darwin-arm64` on port 8083 with `SCRAPERGO_RECOVERY_TRACE=1` and its existing narrow Cloudflare tunnel. Local `diagnostics/recovery.jsonl` and `.recovery-key` are private, Git-ignored, and persist across restarts; do not reset, delete, or upload them.

**Only active generalized duplicate-text tracker:** [ScraperGo Issue #119](https://github.com/richard-mauri/scrapergo/issues/119) and [draft PR #120](https://github.com/richard-mauri/scrapergo/pull/120), on `fix/119-caption-duplication`. The branch name is historical and does **not** limit the issue to captions; retain it because renaming a PR head branch may close the existing PR. **One Issue #119 → one implementation branch → one draft PR #120**. No duplicate-text PRs or issues elsewhere, and no extraction changes in NewsGo. The owner approved diagnostics/investigation, **not** merging the incomplete fix. Do not merge, release, tag, deploy, bump semantic version, delete a branch, or change tokens/tunnels without explicit owner approval. Keep Issue #116 (shared-DOM/performance) deferred.

Start a fresh session by reading `AGENTS.md`, this file, `ARCHITECTURE.md`, and `DEVELOPMENT.md` on the **current PR branch**; check actual GitHub PR/head/CI. For source changes, acknowledge the repository's workflow and safeguard existing tests, secure fetcher/access restrictions, document semantics, Mac/Safari validation, and owner merge gate.

## Defect and exact 100-URL evidence

Original owner-reported defect: ScraperGo returns duplicated article prose and captions, visible through NewsGo. Triggering URL: `https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/`. ScraperGo owns extraction. A separate reusable NewsGo CLI (`~/projects/newsgo/cmd/duplicateaudit`), permanently merged via NewsGo PR #88, audits final `/api/read` JSON; it **does not** modify either server. The URL export on the owner's Mac is `~/Downloads/newsgo-scrapergo-urls.json`.

Historical 100-article comparisons (same exported cohort; all reads succeeded):
- Diagnostic baseline: **106** candidate pairs / **30** flagged articles; this baseline was not conclusively verified as unmodified `main`.
- Preliminary PR #120 caption reconciliation: **97** candidates / **27** flagged articles.
- Latest owner run on new provenance build, Nov 11: **100/100** reads successful, **0** failures, **27** flagged, **100** candidate pairs. Owner's full CSV was uploaded in chat: `duplicate-audit-pr120-provenance.csv` (Mac path `~/Downloads/duplicate-audit-pr120-provenance.csv`). Previous candidate result `duplicate-audit-next.csv` was 97/27, but article content can change between pulls; differences need pairing by URL/source snapshot before attributing causality. There is **no demonstrated generalized improvement** yet.

CSV breakdown, verified from owner-uploaded CSV: **60 exact paragraph/paragraph**; **10 exact image/image**; **2 image/figcaption exact**; **8 paragraph/paragraph overlaps**; **1 heading/paragraph overlap**; **2 image/image overlaps**; **17 paragraph/image overlaps**. Exact and overlap findings are *screening candidates*, not necessarily incorrect duplication. Do not equate matching image captions with duplicate photographs without comparing URLs/source figures. The 100 pairs are concentrated in:
- **LWN article #25: 22 exact p/p**, sampled matched snippets are comment metadata headers (not verified duplicate main-story prose); source DOM/comment handling needs separate examination. Avoid publishing full LWN SubscriberLink access URL.
- **New Scientist paleo diet #10: 15 exact p/p** substantial editorial prose; source/Readability/recovery comparison high priority.
- **Maya site #11: 10** primarily image-caption correlations, may be legitimate reused media text.
- **Ars Polestar #60: 5**, media and paragraph matches.
- **New Scientist quantum error-correction #2: 4 exact p/p** substantial editorial prose; secondary high-priority reproduction.
- **Ars M6 review #1: 4**, including exact p/p caption credit at blocks 4→5 and p/p overlaps at 17→19, 22→24, 25→27. Original defect remains.
- Other examples: voltage-glitching blog #13 four p/p, DeepMind AlphaProof #27 four mixed media/prose. CSV block indices, kinds, and samples are available in the uploaded artifact; these are not source-DOM provenance.

## Recovery provenance — completed and crucial

On PR #120, merged current `main` history/tracing/ARM64 launcher and retained the existing provisional `internal/extraction/figure_caption_text.go` / tests. Added **optional count-only instrumentation**, `internal/extraction/duplicate_trace_counts.go` with tests, `RecoveryStageResult` before/after exact/overlap metrics, `readability-render-text` baseline, `caption-reconciliation` stage, and `reader-postprocessing` checkpoint in `internal/bootstrap/service.go`. Added `cmd/duplicateprovenance` (README and tests): `go run ./cmd/duplicateprovenance -last 100` summarizes recent trace records **without article URLs, text, or IDs**. It ignores historical records without new metrics. GitHub PR notes have the implementation and 100-URL collection method.

Owner completed the full 100-URL run with the new build. `duplicateprovenance -last 100` reported:
```text
stage                    reads    exact before→after    overlap before→after
readability-render-text   100         14→14                  12→12
structured-document      100         14→17                  12→12
leading-paragraphs        91         14→14                  12→12
source-tables            100         17→17                  12→12
caption-reconciliation    91         14→14                  12→9
selected-html-links      100         17→17                   9→9
position-figures          91         14→26                   9→34
editorial-continuity      91         26→26                  34→30
institutional-tail       100         29→29                  30→30
heading-anchored-prose   100         29→75                  30→30
readability-structure    100         75→75                  30→30
reader-postprocessing    100         75→72                  30→28
```
The source of most **newly counted candidates** is after Readability: `heading-anchored-prose` **+46 exact**, `position-figures` **+12 exact +25 overlaps**, `structured-document` +3 exact. `caption-reconciliation` removed 3 overlaps, `editorial-continuity` removed 4 overlaps, final assembly removed 3 exact and 2 overlap. Aggregate totals end at **72 exact + 28 overlaps = 100**, matching the CSV; initial baseline was **14 exact + 12 overlaps = 26**. Per-stage aggregate numbers **cannot map individual duplicates to particular article URLs** and **do not prove every candidate is an error**. In particular, image insertion with its caption naturally creates a candidate against a separately represented caption; heading restoration might insert an actual source paragraph or duplicate an already represented one.

## Source code inspection and next engineering work

Inspected `internal/extraction/heading_prose.go` (`RestoreHeadingAnchoredProse`) and `internal/extraction/figures.go` (`PositionArticleFigures`), plus their tests. `RestoreHeadingAnchoredProse` builds an `existing` normalized-text set of retained blocks, finds ≥3 ordered headings in original HTML, and may recover source paragraphs between them or before the first heading; already protects exact present strings, but may admit normalized/partial/structural variants or source repetitions. `PositionArticleFigures` adds `img` blocks using selected-HTML figure caption text and original placement; may produce candidate image/prose overlap even when both blocks are semantically valid. Existing caption reconciliation only targets figure captions found in selected Readability HTML and has not fixed the original Ars/paleo cases.

**Do not implement a blanket exact-text, prefix/suffix, or global paragraph dedup filter.** Required general invariant: preserve each intentional editorial occurrence in source reading order while preventing reinsertion of the *same* source material through multiple extraction/recovery representations. Protect legitimate repetitions, quotations, comments and live updates, headings/time, inline links, image URLs/captions, tables, bylines, ordering and final word counts.

**Immediate next actions within existing PR #120:**
1. Obtain permitted **source-backed evidence** for New Scientist paleo-diet (#10), New Scientist quantum (#2), and Ars M6 (#1), with the established **secure** `go run ./cmd/htmlcapture -url '...' -out /tmp/<unique>.html` and offline `go run ./cmd/readabilityprobe -input /tmp/<unique>.html -url '...' -out /tmp/<unique>-analysis`; see `cmd/readabilityprobe/README.md`. If publisher blocks capture, report it; do not bypass protections or upload full copyrighted captures. A source capture and Readability `RenderText`/`RenderHTML` comparison is **not yet done**.
2. For each confirmed problematic article, compare original source DOM occurrence/order, selected Readability HTML, `RenderText`, actual canonical blocks, and stage additions. Find whether duplicates are in source, Readability, heading recovery, figure insertion, or downstream reader assembly. Source provenance may need a temporary, opt-in local probe that reveals stage/block details **only on owner's machine**; keep persistent reports private and content-free.
3. Implement **targeted source-backed fix** on the existing branch, with positive/negative synthetic tests: ordinary prose reinsertion, partial overlapping text, caption and credit, same-caption different images, intentional repeated editorial/quoted prose, live updates, LWN-style comment metadata, ambiguous DOM and source-order integrity. Inspect final `article.Blocks`, not just diagnostic counters; preserve security/architecture contracts.
4. GitHub CI on **exact current PR head** (`go test -count=1 ./...`, `go build ./...`), then owner Mac native build, publisher/Safari content/media/links and Print/PDF tests. Rerun the **same** 100 URLs and compare success counts, flagged articles, per-URL findings and content preservation. No claim of success until the original Ars and New Scientist cases improve without legitimate text loss.

## Current validation and checkouts

At handoff start, PR #120 was **draft/open**, tracked latest source head `46faeb8cf89e5b6e44825ea74d3eb17b946e2b41` (this document update creates a newer SHA). GitHub Actions [run 43050231899](https://github.com/richard-mauri/scrapergo/actions/runs/43050231899) passed Go tests/build on that exact preceding head. The owner has already run the ARM64 branch, 100/100 production reads, and new provenance summary successfully. **Not yet verified:** new document-head CI, real publisher source/Readability captures, extraction semantics after a generalized fix, and comprehensive Safari/Print/PDF acceptance. PR must remain draft.

Mac acceptance commands, after confirming no uncommitted files will be overwritten:
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
From separate Terminal, repeat batch only when a new candidate fix is ready; change output filename each run:
```sh
cd ~/projects/newsgo
go run ./cmd/duplicateaudit \
  -input ~/Downloads/newsgo-scrapergo-urls.json \
  -scrapergo https://scrapergo.themauris.org/api/read \
  -public -limit 100 \
  -output ~/Downloads/duplicate-audit-pr120-followup.csv
cd ~/projects/scrapergo
go run ./cmd/duplicateprovenance -last 100
```
Avoid unrelated reads during provenance collection. Confirm original Ars, both New Scientist examples and unrelated articles in Safari (paragraph sequence, headings, links, captions/images, extension handoff, Print/PDF); do not mistake Linux CI for Mac/Safari acceptance.

At meaningful next transition, update **this same `CURRENT_WORK.md`** and PR #120 body to reflect actual head, evidence, remaining blockers and next actions. Do not fork handoff documents.
