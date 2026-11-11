# CURRENT WORK — ScraperGo handoff (November 11, 2026)

## Authoritative state and approval boundary

GitHub `main` is the authoritative **merged** baseline at `41061a1feb9047419b6a0610eaa6c310113f6b30`. PR #122 (native ARM64 launcher) and PR #118 (persistent opt-in recovery diagnostics) were owner-authorized and merged; Issues #121 and #117 were closed. The Mac launcher now runs `dist/scrapergo-darwin-arm64` on port 8083 with `SCRAPERGO_RECOVERY_TRACE=1` and its existing narrow Cloudflare tunnel. Local `diagnostics/recovery.jsonl` and `.recovery-key` are private, Git-ignored, and persist across restarts; do not reset, delete, or upload them.

**Only active generalized duplicate-text tracker:** [ScraperGo Issue #119](https://github.com/richard-mauri/scrapergo/issues/119) and [draft PR #120](https://github.com/richard-mauri/scrapergo/pull/120), on `fix/119-caption-duplication`. The branch name is historical and does **not** limit the issue to captions; retain it because renaming a PR head branch may close the existing PR. **One Issue #119 → one implementation branch → one draft PR #120**. No duplicate-text PRs or issues elsewhere, and no extraction changes in NewsGo. The owner approved diagnostics/investigation, **not** merging the incomplete fix. Do not merge, release, tag, deploy, bump semantic version, delete a branch, or change tokens/tunnels without explicit owner approval. Keep Issue #116 (shared-DOM/performance) deferred.

Start a fresh session by reading `AGENTS.md`, this file, `ARCHITECTURE.md`, and `DEVELOPMENT.md` on the **current PR branch**; check actual GitHub PR/head/CI. For source changes, acknowledge the repository's workflow and safeguard existing tests, secure fetcher/access restrictions, document semantics, Mac/Safari validation, and owner merge gate.

## Current stage: 1/3 source-element provenance (November 11)

**Owner-directed current work:** Stage 1 only of the agreed three-stage Issue #119 approach. Stages 2 (source-backed production correction) and 3 (full 100-URL/Safari validation) are not authorized by this iteration and have not begun. Separate new [Issue #123](https://github.com/richard-mauri/scrapergo/issues/123) proposes a future staging Render/CI/agent-assisted loop; it is not deployed or part of this PR.

**What is established:** ScraperGo's main extraction initially splits Go Readability `RenderText()` into ordered paragraph chunks, before source HTML recovery stages. New Scientist paleo same snapshot improved from 15 exact pairs / 92 blocks to 0 pairs / 77 blocks after the canonical `RestoreHeadingAnchoredProse` comparison. Full 100-URL audit improved from 100 candidates/27 articles to **54 candidates/21 articles**, with 100/100 successful reads; article-content preservation remains unverified. For original Ars M6, private capture `/tmp/pr120-ars-m6.html` (SHA256 `8bc65dad6714447409e5a6a3b6367339b0363f86f0703d4e6e500eb0a8055b9a`), four pairs appear in the **initial Readability text**: 4/5 exact; 17/19, 22/24, 25/27 overlaps. No recovery stage changed them. A prior source-evidence search found source minimal tags `div:1,span:2` for each overlap pair, while Readability returned `p:1,span:1` minimal tags and two containing paragraphs; pair 4/5 had no minimal-source match. These are suggestive counts, **not** a complete source-node provenance.

**Stage 1 implementation added on existing PR #120 branch:** `cmd/duplicateprobe/source_provenance.go`, `source_provenance_test.go`, and a `-source-provenance` switch in `cmd/duplicateprobe/main.go`. This opt-in, network-free diagnostic inspects **both members of each candidate pair** in the same private captured HTML and independently rendered Readability-selected HTML. For each matching **minimal** source element, it reports a document-order ordinal, index-qualified element path, immediate parent tag, nearest semantic container (article/section/figure/figcaption/etc.), origin (element text versus safe-listed attribute name such as image `alt`), and text-versus-weaker whitespace-compact matching mode. It prints no publisher text, URL, attribute values, IDs, classes or HTML and writes no files. Multiple source occurrences and independently selected paragraphs can be compared in order, but the tool cannot prove one-to-one transformation because Readability does not expose original node IDs. Synthetic tests check two independent paragraphs, one source container versus two selected paragraphs, figure caption plus image alt, inline text, and noise exclusion. Production extraction unchanged.

**Next action — owner must run on saved private Mac capture, no download:** After pulling this branch and running `go test -count=1 ./...`, execute:

```sh
cd ~/projects/scrapergo
go run ./cmd/duplicateprobe \
  -input /tmp/pr120-ars-m6.html \
  -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/' \
  -source-provenance
```

Paste only stdout. Compare ancestry/order for p4/5 and each overlap to determine if source offers one element, multiple independent elements, media alt/credit, or ambiguous transformed content. If ambiguous, keep staging investigation; **do not manufacture a source/Readability mapping or remove text merely because it overlaps**. Stage 2 begins only with a justified general source-semantic invariant plus positive/negative tests. Current PR must remain draft, no owner production/Safari merge acceptance yet. Document exact-head CI status after final documentation commit.

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

## New offline article-level source probe (November 11 follow-up)

A new `cmd/duplicateprobe` command and unit tests now live **on this same draft PR #120**. It accepts a private saved HTML capture (`-input`) plus the source URL (`-url`), calls ScraperGo's existing `bootstrap.ExtractWithEvidenceTraced` and final `reader.DefaultProcessor` offline (no publisher request), prints per-stage candidate counts and final matching canonical block indices/kinds with a short digest, and **does not print or persist article text**. It does not change normal server extraction. Read `cmd/duplicateprobe/README.md` for exact secure capture/probe commands for New Scientist paleo-diet and Ars M6; `cmd/readabilityprobe` still supplies the comparison of source and Readability HTML/text. If secure `htmlcapture` fails on a paywall/challenge, report the limitation rather than bypassing it. **This source capture and the offline probe have not yet been run on the owner's Mac.** This tool is an investigation aid, not the generalized fix. The public Ars page's text view showed multiple caption/credit repetitions, but web text rendering is not a reliable raw-source DOM capture; do not conclude source vs extraction provenance without owner-side source snapshot.

## Source-backed New Scientist paleo reproduction and first scoped fix (November 11, 2026)

The owner supplied a Mac transcript `typescript(20261111-164336)`. Secure `cmd/htmlcapture` downloaded the public paleo-diet URL `https://www.newscientist.com/article/2598048-a-true-paleo-diet-would-include-a-lot-more-insects/` with HTTP 200; **273,809 bytes**, SHA-256 `b50af8a3c3f4d8f39477fb03694455b9b48b4b8b4f03886a0499fd4bf2c266a1`, saved privately on owner Mac at `/tmp/pr120-paleo.html`. `readabilityprobe` saw title “A true paleo diet would include a lot more insects”, byline Michael Marshall, and selected-HTML headings “Crunchy” and “Om nom nom”; raw source has many additional navigation headings. Do not assume the developer has access to the raw Mac capture; only the terminal transcript was uploaded.

**Actual offline pipeline evidence** on that snapshot: `go run ./cmd/duplicateprobe -input /tmp/pr120-paleo.html -url 'https://www.newscientist.com/article/2598048-a-true-paleo-diet-would-include-a-lot-more-insects/'` reported **92 blocks and exactly 15 duplicate paragraph pairs**. `readability-render-text` was 0 exact/0 overlap; `structured-document`, `source-tables`, `selected-html-links` also stayed at zero; `heading-anchored-prose` went **0→15 exact**; `readability-structure` and `reader-postprocessing` remained at 15. Example 1-based matching block pairs: 1/7, 2/8, 3/10, 14/20, 15/23, 37/53. This directly demonstrates that `RestoreHeadingAnchoredProse` reintroduces the duplicated prose **for this particular real article**; do not infer all other flagged articles share the origin.

**Narrow source-backed hypothesis and candidate fix now committed on PR #120:** `RestoreHeadingAnchoredProse` compared `sourcehtml.Normalize(sourcehtml.Text(n))` with already selected canonical `article.Block.Text`. `sourcehtml.Text` adds whitespace between descendant inline nodes; `structured.TextBlock` joins actual text runs, meaning semantically identical styled/link paragraphs can evade the previous comparison. `internal/extraction/heading_prose.go` now checks the **canonical `structured.TextBlock` text** against retained paragraphs in both introduction and inter-heading recovery. Added `internal/extraction/heading_prose_canonical_test.go` with styled source paragraphs already retained; prior positive recovery fixtures stay intact. This is *not* a global deduplication stage, and **the owner's same-snapshot retest has not occurred**; there is no proof yet that it removes all 15 live pairs without collateral loss. Check PR head/CI after these commits.

**Immediate next verification — no recapture required:** On Mac, update branch/build, then rerun `go run ./cmd/duplicateprobe -input /tmp/pr120-paleo.html -url 'https://www.newscientist.com/article/2598048-a-true-paleo-diet-would-include-a-lot-more-insects/'` against that **unchanged snapshot**, verifying previously 0→15 heading-recovery additions have declined and canonical prose/headings remain. If no improvement, don't broaden blindly: inspect actual source nodes/selected block membership locally (private, opt-in) and revise within same PR. Then test Ars source snapshot independently, rerun the same 100-URL NewsGo batch, compare per-URL candidates, check Safari/print and obtain owner approval before merge.

## Owner same-capture acceptance result — New Scientist paleo (Nov 11)

After pulling the narrow canonical-text comparison fix on PR #120, owner reran `go run ./cmd/duplicateprobe -input /tmp/pr120-paleo.html -url 'https://www.newscientist.com/article/2598048-a-true-paleo-diet-would-include-a-lot-more-insects/'` **without recapturing HTML**. **Before: 92 blocks and 15 exact pairs, originating entirely in heading-anchored-prose recovery. After: 77 canonical blocks and zero exact/overlap candidate pairs.** All listed stage counts, including heading recovery and final reader postprocessing, were 0/0; extraction method remained Go Readability v2. This directly validates the narrow fix against the previously failing source snapshot and removes 15 previously duplicated blocks. It does **not** independently demonstrate content preservation; inspect real paragraph/heading/link/photo content and Safari Print/PDF.

The owner-side result has been added to Issue #119. **Next test:** use the existing `cmd/htmlcapture`, `cmd/readabilityprobe` and `cmd/duplicateprobe` against the original Ars M6 article (documented in `cmd/duplicateprobe/README.md`), then rerun the exact same NewsGo 100-URL audit with a **new output filename** and `go run ./cmd/duplicateprovenance -last 100`. Historical latest 100-URL candidate benchmark before this fix is **100 matches across 27 flagged articles**, 100/100 success. Do not claim generalized resolution solely from the paleo success; scrutinize remaining image-caption and comment-metadata matches separately. Keep draft PR #120 unmerged pending owner acceptance.

## Same 100-URL owner regression after canonical heading fix (Nov 11, 2026)

Owner uploaded `duplicate-audit-pr120-heading-fix.csv` and Terminal transcript `typescript(20261111-165356)`. NewsGo `cmd/duplicateaudit` replayed the same 100 exported URLs through `https://scrapergo.themauris.org/api/read`: **100/100 successful, 0 failed, 21 flagged articles, 54 duplicate candidates**. Previous run of the same cohort before the heading fix: **100 successful, 27 flagged, 100 candidate pairs**. Thus **46 fewer screening candidates (−46%) and six fewer flagged articles**. CSV comparisons by article index and URL: 70→24 exact matches, 17→17 `image_caption_or_text_overlap`, 11→11 `caption_or_text_overlap`, 2→2 `image_exact`. The entire measured reduction is in exact matches.

Key per-article reductions:
- LWN article #25: **22→0** exact pairs, previously comment-metadata headers; verify preserved reader comments/legitimate prose, not merely loss of content.
- New Scientist paleo #10: **15→0** exact prose; separately confirmed by same-snapshot offline probe, 92→77 blocks and all stage duplicate counts 0 after heading recovery change.
- New Scientist quantum #2: **4→0** exact prose.
- VulnCheck #32: **2→0**; DeepMind AlphaProof #27: 4→3; Claude #31 and Autopian #20: 1→0 each.
- **Original Ars M6 #1: unchanged 4→4** (photo-credit exact p/p and three caption/text p/p overlaps).
- **Maya site #11: unchanged 10→10** predominantly media/caption overlap or repeated same-caption images.
- Other unchanged concentrations: Ars Polestar #60 5, voltage-glitching #13 4, Ars drone article #61 4, Ars climate #55 3, abandoned architecture #33 3.

The owner transcript contains the 100-URL CLI and an offline paleo `duplicateprobe` reiterating 77 blocks/0 pairs, but **does not contain a new `go run ./cmd/duplicateprovenance -last 100` report**. No independent content-preservation or Safari Print/PDF acceptance yet. The 100 URL statuses remained HTTP 200; this does not prove each source paragraph was preserved. Captions/credits can legitimately duplicate text and same-caption image matches may represent different URLs. Do **not** blindly delete identical strings or remove the figure insertion stage.

**Next specific work:** Investigate original Ars M6 (#1) using secure `htmlcapture` and offline `readabilityprobe`/`duplicateprobe` commands in `cmd/duplicateprobe/README.md`; compare exact source figure/caption placement, preserved prose and image URLs in final output; add positive/negative regression fixtures for evidence-backed text overlap only. Check content preservation on New Scientist and LWN, and run `duplicateprovenance -last 100` if no unrelated article reads have displaced the batch. Keep Issue #119 / draft PR #120 as the only tracker. The next repair must preserve existing semantic content and be tested against the identical 100 URLs. **Do not merge or deploy without explicit owner authorization.**

## Original Ars M6 fixed-snapshot diagnosis — November 11, 2026

Owner successfully acquired original Ars M6 review using the **existing** validated `cmd/htmlcapture` boundary (HTTP 200, HTML; **212,564 bytes**, SHA-256 `8bc65dad6714447409e5a6a3b6367339b0363f86f0703d4e6e500eb0a8055b9a`). Private capture exists on owner Mac at `/tmp/pr120-ars-m6.html`; **do not commit or upload the full source**. Reproduction URL: `https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/`.

Owner's offline `go run ./cmd/duplicateprobe -input /tmp/pr120-ars-m6.html -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/'` reported **68 canonical blocks**, method **Go Readability v2**, and **four candidate pairs**: block **4(p)/5(p)** exact (hash `22f6070e195b`); **17(p)/19(p)** overlap (`5139e673818d`); **22(p)/24(p)** overlap (`a216b793c458`); **25(p)/27(p)** overlap (`6b8d8e370523`). Crucially, `readability-render-text` starts at **1 exact / 3 overlaps**, and **every subsequent stage** (structured document, leading, figures, editorial continuity, heading-anchored, final reader) shows the same **1 / 3** before and after. Therefore the **heading-recovery fix cannot eliminate these pairs**, and ScraperGo's later stages are not responsible for introducing them. **Still unknown:** whether source HTML intentionally has two textual representations (captions, credits, hidden/alternate media), or whether Readability `RenderText` itself duplicates them. The fact that they predate recovery does **not** prove Readability is defective.

**Next action before any Ars production code change:** Run `cmd/readabilityprobe` using **the unchanged saved** `/tmp/pr120-ars-m6.html` with the original URL and a unique *local private* output directory, then inspect the selected `readability.html` and `readability.txt` against `source.html` specifically around the four block pairs, checking DOM ancestry, figure/figcaption/credit markup, and whether the same words really occur twice in the source. Use `umask 077` before the command because `readabilityprobe -out` creates files that include copyrighted full article HTML; keep all raw source and HTML local. See `cmd/readabilityprobe/README.md`. If the source contains the repetitions as independent editorial occurrences, preserve them; if Readability flattens one figure/credit into multiple textual copies, implement a publisher-independent, source-backed reconciliation (with positive/negative tests) on **existing PR #120 only**. Do not blanket-deduplicate paragraphs or filter based on publisher/URL.

**Latest broad benchmark:** 100/100 successful URL reads, 21 flagged, **54 candidates**, after scoped heading correction. Ars remains **4**. New Scientist paleo same source capture **15→0** candidates with 92→77 blocks, and 100-URL batch reduced **100→54** candidates. Content-preservation/Safari and Print/PDF checks not completed. Issue #119 and PR #120 remain open/draft; no merge/release/deploy approval.

## Ars Readability structure report + optional source-element evidence (Nov 11)

Owner ran `cmd/readabilityprobe` on the **unchanged** `/tmp/pr120-ars-m6.html` source capture. Report: title "Review: Apple's M6 MacBook Pro is good, but it's missing all the exciting stuff", byline Andrew Cunningham. Source HTML contains one article and one section; h2 "Look and feel", "Performance and power: M6 redux", "Pro-ish", h3 "The good", "The bad". Readability-selected HTML preserves those five headings and order and includes one article, zero sections; source `time` "Nov 10, 2026 7:00 am" / `2026-11-10T07:00:21-05:00` did not survive into selected HTML. **This report enumerates headings/timestamps, not paragraph and caption duplication.** Prior offline `duplicateprobe` still establishes 1 exact + 3 overlaps already at initial Readability RenderText; it does not distinguish duplication in original source from Readability flattening.

**Added diagnostic-only follow-up** on existing PR #120: `cmd/duplicateprobe -source-evidence` uses the saved capture to count how many original-source vs Readability-selected HTML `p` and `figcaption` elements contain each candidate's shorter matched text, keyed by 1-based canonical block pair. It **does not print article text or write source files**. New `cmd/duplicateprobe/source_evidence.go` and tests, and README instructions. This is a conservative text-presence heuristic, not exact identity of DOM occurrences; inline whitespace can cause misses and intentional repetition remains possible. No production extraction code change. To run after pulling current branch, use:

```sh
cd ~/projects/scrapergo
go run ./cmd/duplicateprobe \\
  -input /tmp/pr120-ars-m6.html \\
  -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/' \\
  -source-evidence
```

Inspect these per-pair source/selected element counts first. Determine whether repeated text represents independent source paragraphs, figure captions/credits or duplicated Readability rendering; only then design a source-backed fix and preserve negative tests. **No global removal**, new Issue, PR, merge or release.

## Ars source-vs-Readability element evidence; expanded offline matcher (Nov 11, 2026)

Owner ran the PR #120 `cmd/duplicateprobe -source-evidence` against the unchanged Ars M6 capture `/tmp/pr120-ars-m6.html` (212,564 bytes, SHA-256 `8bc65dad6714447409e5a6a3b6367339b0363f86f0703d4e6e500eb0a8055b9a`). Pipeline remains **68 canonical blocks, 4 pairs**: p4/p5 exact, p17/p19, p22/p24, p25/p27 overlaps. **All four** already appear at initial `readability-render-text`; none is inserted by later recovery stages. For **every** pair, the narrow source match counted **0 source `p` / 0 source `figcaption`** and **2 Readability-selected `p` / 0 selected `figcaption`**. This establishes that the selected HTML has two paragraph representations containing each matched passage; it **does not** establish whether Readability duplicated the content from another source tag, because source markup may use other containers, nested inline text, or different whitespace serialization.

**Next diagnostic committed on the same PR:** Extended `cmd/duplicateprobe -source-evidence` to report `Source minimal tags` and `Readability minimal tags`, counting smallest matching HTML elements across source tags (including `div` and inline structures) while excluding blocked tags and preventing ancestor double counting; added tests for inline markup, `div`, `figcaption`, and scripts. This is an offline investigative extension, **not a production extraction change**, and still does not guarantee semantic identity. README describes limitations. Owner should pull the current branch and rerun the same saved Ars capture with `-source-evidence` (no new download). Compare tags: two distinct source matches might be genuine authorial repetition or duplicate media/credit display; one source match and two Readability-selected matches suggests duplication in selection/transformation. If no source match, inspect source locally without exposing full copyrighted HTML. Do not globally strip repeated text.

## Latest Ars source-element attribution (owner Mac, November 11)

Owner reran the **unchanged** private `/tmp/pr120-ars-m6.html` capture with the expanded `cmd/duplicateprobe -source-evidence` at PR #120 head `c5ac3ab8bf85edf1eae2757e05fe0e2455f75a40`. Result remains **68 blocks, 4 candidate pairs** (1 exact, 3 overlaps), all **already present at `readability-render-text`** and unchanged by every subsequent stage. Source/selected counts per pair:

```text
pair   source p/caption   selected p/caption   source minimal tags   selected minimal tags
4/5         0/0                 2/0                   none                    none
17/19       0/0                 2/0                   div:1,span:2            p:1,span:1
22/24       0/0                 2/0                   div:1,span:2            p:1,span:1
25/27       0/0                 2/0                   div:1,span:2            p:1,span:1
```

**Interpretation, with limits:** Readability-selected HTML contains two p elements associated with each shorter text, while the original source represents three overlapping passages using nested div/span markup. Minimal-source matching counts can double-count overlapping sibling/ancestor representations or miss exact matches due to whitespace/HTML entities. Do **not** claim Readability invented the duplicates solely because p/figcaption count is zero. Pair 4/5 has no smallest-element hit despite two selected paragraphs, so its provenance is still unknown. This result makes Readability selection/flattening a stronger investigative target than ScraperGo recovery, but is not enough to authorize a deletion rule.

**Next evidence gate:** Use an *offline*, privacy-preserving local DOM inspection/probe that reports original element ancestry and source ordering for each candidate in the saved HTML, and whether selected p duplicates come from one source node or multiple independent source occurrences. Inspect figure/credit semantics and any alternate/hidden markup. Only then write a generalized source-backed extraction repair with negative tests for intentionally repeated paragraphs, same caption on different photos and reader content preservation. The 100-URL after-heading baseline remains **100 successful / 21 flagged / 54 candidates**. Keep single Issue #119 / draft PR #120, no merge/deploy/release without owner approval.

## Source code inspection and next engineering work

Inspected `internal/extraction/heading_prose.go` (`RestoreHeadingAnchoredProse`) and `internal/extraction/figures.go` (`PositionArticleFigures`), plus their tests. `RestoreHeadingAnchoredProse` builds an `existing` normalized-text set of retained blocks, finds ≥3 ordered headings in original HTML, and may recover source paragraphs between them or before the first heading; already protects exact present strings, but may admit normalized/partial/structural variants or source repetitions. `PositionArticleFigures` adds `img` blocks using selected-HTML figure caption text and original placement; may produce candidate image/prose overlap even when both blocks are semantically valid. Existing caption reconciliation only targets figure captions found in selected Readability HTML and has not fixed the original Ars/paleo cases.

**Do not implement a blanket exact-text, prefix/suffix, or global paragraph dedup filter.** Required general invariant: preserve each intentional editorial occurrence in source reading order while preventing reinsertion of the *same* source material through multiple extraction/recovery representations. Protect legitimate repetitions, quotations, comments and live updates, headings/time, inline links, image URLs/captions, tables, bylines, ordering and final word counts.

**Immediate next actions within existing PR #120:**
1. Obtain permitted **source-backed evidence** for New Scientist paleo-diet (#10), New Scientist quantum (#2), and Ars M6 (#1). First run the **exact** secure capture / offline Readability / new `duplicateprobe` commands in `cmd/duplicateprobe/README.md` for paleo and Ars. Use the established **secure** `go run ./cmd/htmlcapture -url '...' -out /tmp/<unique>.html` and offline `go run ./cmd/readabilityprobe -input /tmp/<unique>.html -url '...' -out /tmp/<unique>-analysis`; see `cmd/readabilityprobe/README.md`. If publisher blocks capture, report it; do not bypass protections or upload full copyrighted captures. A source capture and Readability `RenderText`/`RenderHTML` comparison is **not yet done**.
2. For each confirmed problematic article, compare original source DOM occurrence/order, selected Readability HTML, `RenderText`, actual canonical blocks, and stage additions. Find whether duplicates are in source, Readability, heading recovery, figure insertion, or downstream reader assembly. Source provenance may need a temporary, opt-in local probe that reveals stage/block details **only on owner's machine**; keep persistent reports private and content-free.
3. Implement **targeted source-backed fix** on the existing branch, with positive/negative synthetic tests: ordinary prose reinsertion, partial overlapping text, caption and credit, same-caption different images, intentional repeated editorial/quoted prose, live updates, LWN-style comment metadata, ambiguous DOM and source-order integrity. Inspect final `article.Blocks`, not just diagnostic counters; preserve security/architecture contracts.
4. GitHub CI on **exact current PR head** (`go test -count=1 ./...`, `go build ./...`), then owner Mac native build, publisher/Safari content/media/links and Print/PDF tests. Rerun the **same** 100 URLs and compare success counts, flagged articles, per-URL findings and content preservation. No claim of success until the original Ars and New Scientist cases improve without legitimate text loss.

## Current validation and checkouts

At the prior handoff, PR #120 was **draft/open**, tracked previous source head `46faeb8cf89e5b6e44825ea74d3eb17b946e2b41` (subsequent diagnostic-tool commits and this document update create newer SHAs; inspect actual current GitHub PR head). Prior GitHub Actions [run 43050231899](https://github.com/richard-mauri/scrapergo/actions/runs/43050231899) passed Go tests/build on that exact preceding head. The owner has already run the ARM64 branch, 100/100 production reads, and new provenance summary successfully. **Not yet verified:** CI at the latest offline-probe head, real publisher source/Readability captures, extraction semantics after a generalized fix, and comprehensive Safari/Print/PDF acceptance. PR must remain draft.

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
