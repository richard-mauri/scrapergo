# Issue #125 — Structure-preserving article extraction evaluation

**Status:** Phase 1 design and baseline inventory only. No production behavior change and no switch away from Readability.

## Question

Can a sanitized, source-DOM-based article representation preserve more editorial meaning (especially photograph-caption-credit relationships and document order) while filtering advertisements, overlays and repeated responsive markup, **without** weakening ScraperGo's current extraction completeness, security or access controls?

The current merged implementation is the benchmark to **beat**, not a disposable prototype.

## Verified baseline (November 11, 2026)

- Authoritative starting `main`: `2832ac7cde154af2910042cd030feb67f9f905d3`.
- [PR #120](https://github.com/richard-mauri/scrapergo/pull/120) merged generalized duplicate-text improvements. The owner Mac's **unchanged private** Ars M6 HTML snapshot went from 68 blocks / 3 overlapping caption pairs to 64 blocks / 0 overlapping pairs after caption reconciliation. One exact candidate remains. The owner Safari screenshots also show a still-repeated hero caption and photograph captions detached from their own paired photographs. These **do not prove** a new DOM extractor will be better. The remaining focused defect is [#124](https://github.com/richard-mauri/scrapergo/issues/124).
- Earlier New Scientist paleo report: 92 blocks / 15 screening candidates → 77 blocks / zero after heading-anchored correction. Prior 100-URL cohort: 100/100 returned successful reads, with 54 screening candidate pairs across 21 articles; **not** yet a final same-input test of the current merged extractor. Counts are screening signals, not a losslessness proof.
- The source capture `/tmp/pr120-ars-m6.html` remains **private on the owner's Mac**; its publisher HTML, user screenshots, credentials and diagnostics are not part of this PR.

## What the code currently does

`internal/extraction/readability_pipeline.go` calls `readability.FromReader` on already retrieved HTML (no new network call), then `parsed.RenderText`. It splits the text on blank lines and initially creates `article.Block{Kind:"p"}` for every accepted chunk. Those blocks pass through source-backed recovery stages including structured-document recovery, leading paragraphs, source tables, caption reconciliation, Readability-selected links, `PositionArticleFigures`, continuity/tail/heading recovery, and selected-HTML structural restoration. `parsed.RenderHTML` is separately used as evidence for link, figure, caption and heading reconstruction. The existing legacy and partial-HTML fallback paths are preserved.

This identifies a **potential loss-of-structure boundary at `RenderText` / splitting into paragraphs**, even though later stages may recover reliable semantics. It does **not** establish that the original source DOM presents a unique visible representation: responsive alternatives can duplicate captions, hidden branches may be ambiguous, and browser-visible CSS state is not fully described by fetched HTML.

`ARCHITECTURE.md` allocates HTML parsing to `internal/sourcehtml`, extraction to `internal/extraction`, canonical transport-independent blocks to `internal/article`, media identity to `internal/media`, reader stages to `internal/reader`, and UI presentation to `internal/ui`. Retrieval belongs to `internal/fetcher` and related established security-controlled boundaries. A prototype must respect these boundaries.

## Proposed experiment — not yet implemented

1. **Article scope selection.** Use Readability's article selection as supporting evidence, then select or clone one source DOM article region without serializing it to plain text first. Reject uncertain or non-editorial scopes conservatively; preserve fallback.
2. **Structural editorial filtering.** Exclude noneditorial navigation, adverts, interstitials, popups, newsletter prompts, trackers, and interactive controls using semantic and structural evidence, not blanket keyword deletion. Detect alternative/responsive representations without collapsing genuinely separate photos or repeated editorial paragraphs. Do not assume a lack of `hidden` attributes proves an element is visible.
3. **Safe canonicalization.** Produce source-ordered, sanitized semantic objects: headings, prose and inline links, lists/tables, single images, and image groups with **explicit image-caption-credit attribution**. Continue to rely on the established secure URL resolution and image delivery; never forward arbitrary publisher scripts, styling, event handlers, active embeds, or unvalidated URLs.
4. **Observable fallback.** When article scope or image ownership is ambiguous, retain the currently accepted Readability output or record a noncommittal diagnostic result. Do not silently discard blocks or invent relationships. No network requests in extraction or rendering stages.

No claim is made that existing `article.Block` can represent every required relationship; evaluate an **internal-only** group or association model before proposing canonical model/API changes. Do not ship raw HTML directly to the browser.

## Reproducible comparison contract

Run **both implementations on identical captured HTML** and compare final document semantics, not only text counts:

| Dimension | Evidence |
| --- | --- |
| Editorial completeness | Exact missing/extra paragraphs, independent intentionally repeated text, headings and timestamps |
| Structural order | Source-backed block order; nesting and grouping of figures, lists, tables, and prose |
| Media identity | Number and identity of distinct editorial images; no duplicate responsive assets |
| Figure ownership | Exactly one appropriate caption and credit attached to each photograph, correct left/right gallery order |
| Cleanliness | Remaining ad/popup/navigation/promo elements and mistakenly removed editorial material |
| Security and access | Script/handler/style/embed removal, safe links/media, unchanged denied/partial-page behavior |
| Quality | Side-by-side rendered Safari output and Print/PDF, including links and images |
| Cost | Comparable CPU, allocation, duration and fallback rates on matched fixtures/environment |

Start with deterministic synthetic positive and **negative** fixtures for simple articles, competing main containers, hidden/duplicate responsive gallery variants, two photos with the same caption, unrelated repeated prose, malicious links and scriptable HTML, tables/lists, incomplete/blocked content and ambiguous photo ownership. An isolated test-only prototype must neither alter `ExtractWithReadability` nor the deployed reader. Do not use publisher-specific URL/class gates.

The owner can later run **offline** Ars M6 and New Scientist captures on the Mac using existing tools, followed by the stored 100-URL audit in an approved manner. Publishing private captures or relying on changed live publisher HTML is not acceptable. Synthetic tests cannot substitute for Safari acceptance, and GitHub CI does not run the owner's local publisher snapshots.

## Decision gate and separation of concerns

- **Phase 1 (this PR iteration):** Source inventory, design and comparison criteria. Documentation only; no runtime acceptance claimed.
- **Phase 2 (future iteration in this same draft PR, if approved):** Isolated, opt-in, nonproduction prototype with positive and negative tests.
- **Phase 3:** Evidence-driven evaluation against the merged Readability path and a written adopt/continue/reject decision. A separate owner-approved implementation plan is required before any production extraction switch.

[#124](https://github.com/richard-mauri/scrapergo/issues/124) remains the focused current UI/media defect; [#116](https://github.com/richard-mauri/scrapergo/issues/116) remains a separate measured shared-DOM performance investigation; [#123](https://github.com/richard-mauri/scrapergo/issues/123) remains staging/agent automation. Do not widen this PR into those issues.

**Safety and release boundary:** No bypass of publisher access restrictions; no changes to private tokens, the Cloudflare tunnel, service deployment, semantic version, production default extractor or release. Do not merge without explicit owner authorization.
