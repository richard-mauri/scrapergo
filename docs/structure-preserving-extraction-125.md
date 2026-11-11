# Issue #125 — Structure-preserving article extraction evaluation

**Status:** Phase 2 initial offline-only proof of concept. No production behavior change and no switch away from Readability.

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

## Implemented first offline prototype (not production)

`internal/sourcepreserve.Inspect` uses Go's existing `golang.org/x/net/html` parser on an already-saved input string. It **requires exactly one nonexcluded `<article>` element**; ambiguous or absent scopes return `status: inconclusive` rather than claiming a correct result. It observes source-ordered heading and paragraph text, validated HTTP(S) links, and simple explicit `<figure>` wrappers with exactly one `<img>` and at most one `<figcaption>`. For captions, optional explicit `<small>` or `<cite>` credit text is separately reported; no credit is inferred from substring patterns or nearby text. When a figure has multiple images, an unsafe URL, or ambiguous caption structure, it increments `ambiguousFigures` and **does not invent photo-to-caption ownership**. It doesn't render HTML, sanitize HTML for browser insertion, or claim that absence of an HTML hidden attribute proves the source branch is browser-visible.

Conservative exclusion removes executable/interactive elements (scripts, styles, iframes, forms, etc.), noneditorial landmark regions (nav, aside, footer), explicit hidden/aria-hidden subtrees, role=dialog/alertdialog and a small set of exact noise class/id tokens. It **does not** implement arbitrary CSS, dynamic publisher layout visibility, an ad blocker, responsive gallery deduplication, all URL policy checks, tables/lists preservation, or source/Readability article-scope matching. Its output is an experimental semantic report, **not** a safe rendered HTML document or canonical `article.Article`. The existing production `internal/extraction` and `internal/reader` packages do not import or call this prototype. The implementation is intentionally conservative and incomplete; the report's skipped/ambiguous counters must be reviewed.

Use it locally without network access:

```bash
cd ~/projects/scrapergo
go test -count=1 ./internal/sourcepreserve ./cmd/structureprobe
go run ./cmd/structureprobe \
  -input /tmp/pr120-ars-m6.html \
  -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/'
```

The JSON report is printed to standard output; **it contains observed article text and image/link URLs**, so do not publish the entire report or private source HTML. There is no network fetch and no file write by this command. We can compare it with the existing offline `cmd/duplicateprobe -source-provenance` for the **same snapshot**, but those instruments have different representations; a claim of superior extraction requires a matched per-image/prose/links audit and later Safari acceptance.

Focused synthetic tests cover ordered paired figures with distinct credits, exact same caption on two distinct images, intentional repeated editorial prose, explicit ad/dialog/hidden removal, unsafe link and image schemes, ambiguous two-image figure, multiple article scopes, and unsafe base URLs. They do not establish real publisher fidelity.

**Next experiment:** Decide how to match a source article container selected by Readability; represent grouped gallery/responsive variants and per-image captions; preserve inline formatting/tables/lists; evaluate secure filtering with truly ambiguous markup; and construct a repeatable baseline comparator before proposing runtime integration. Do not merge or release this proof of concept as a production replacement.

## First owner Ars M6 replay and ambiguity classification (November 11, 2026)

The owner ran `go test -count=1 ./...` on the Mac and **all Go packages passed**. The frozen private `/tmp/pr120-ars-m6.html` produced `status: candidate`, source-ordered headings and many editorial paragraphs with safe links, **one explicit figure** (the image at `/wp-content/uploads/2026/11/IMG_9764.jpeg` reported with its full caption), **19 ambiguous figure/image encounters**, and **28 skipped noneditorial subtrees**. This is a substantive first real-publisher check, but **not** a head-to-head fidelity win over Readability: the source gallery photographs remain unsupported and skipped regions were not audited for editorial losses. The `ambiguousFigures` value is not a count of unique editorial photographs, and `skippedNoneditorial` is not evidence that all skipped content was genuinely unwanted.

The diagnostic now further classifies ambiguous encounters without publisher text: `unwrappedImages` (individual image tags not inside supported `figure`), `multiImageFigures` (more than one image in one figure), `invalidFigureImages` (zero eligible images or unsafe/invalid image URL), and `multipleCaptions` (more than one explicit figcaption in a figure). A single figure may increment more than one reason, so **reason counts are not guaranteed to sum to `ambiguousFigures`**. The existing count remains for compatibility. No source-specific removal, image attribution, gallery association, or production changes have been made. New synthetic tests check classification and that nearby independent prose is preserved.

**Next owner local action:** rerun the unchanged `cmd/structureprobe` invocation after pulling this PR branch and report only `status`, the five ambiguity counts and `skippedNoneditorial`. These fields do not disclose article text or media URLs. Use the breakdown to select the next narrowly justified step: general image wrappers or true multi-image figure structures. Do not infer that skipped images are ads or that adjoining text is a caption.

## Owner's second Ars replay: unsupported image wrappers are the entire current ambiguity (Nov 11)

The owner reran the same private Ars capture at the latest diagnostic build. The Mac full `go test -count=1 ./...` **passed**. Summary: `status=candidate`, `ambiguousFigures=19`, `unwrappedImages=19`, `skippedNoneditorial=28`; the remaining optional ambiguity counters were absent/zero. **All 19 ambiguous encounters arise from `img` elements outside recognized `figure` containers**, not multi-image `figure` elements or unsafe figure sources. This does not establish 19 unique editorial image identities or prove that 28 removed subtrees are noneditorial.

The next offline-only diagnostic therefore adds `compactImageTextWrappers` and `otherUnwrappedImages`. A compact wrapper is counted only when an image has, within at most three ancestor edges below `article`/`main`/`body`, a containing subtree with **exactly one image** and a **separate non-image direct child containing text**. This reports a *structural opportunity*, **not** a confirmed caption, photographer credit, same-photo identity, browser visibility, or permission to delete a duplicate. All remaining unwrapped images increment `otherUnwrappedImages`. These two counts partition the 19 unwrapped-image encounters. Positive/negative synthetic fixtures cover single-image text siblings, a multi-photo gallery with unrelated shared text, and distant editorial paragraphs that must not be attributed to a photograph. No production behavior or even experimental item extraction was changed, only the read-only counters.

**Next owner action:** after pulling the draft PR branch, re-run the same offline `cmd/structureprobe` command piped to `tail -n 12`; share only the final numeric fields, not publisher text/URLs. Then use the classification to design a source-backed per-photo association model and verify the selected example against original publisher Safari. Do not prematurely merge or integrate into production.
