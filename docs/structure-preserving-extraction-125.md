# Issue #125 — Structure-preserving article extraction evaluation

**Status (owner-approved pivot, November 11, 2026):** Stop developing proximity/ancestor-count image-wrapper heuristics. Retain the existing offline inspectors and synthetic tests solely as research evidence. The next active experiment is **structural conversion of Readability-selected HTML**, with original source DOM as a fallback/comparison source. No production behavior change or switch away from Readability.

## Owner-approved architectural pivot: preserve selected HTML as a tree

**Decision:** The previous exploration of 3–4-ancestor searches, 2–8-image containers, compact text siblings and per-image adjacency is too brittle to justify a production extractor. It demonstrated that source HTML contains information lost by flattening, but did **not** prove reliable gallery or caption ownership. **Freeze those diagnostics**: `internal/sourcepreserve` and `cmd/structureprobe` are historical, opt-in offline probes only. Do not add more limits, scoring rules or publisher-specific shortcuts. They are **not** an implementation candidate to merge into production.

### Three parallel inputs, one comparison contract

| Candidate | Responsibility | Scope |
| --- | --- | --- |
| A. Current merged Readability + recovery | Control / production behavior | Run `ExtractWithReadability` and the existing reader's final canonical blocks unchanged |
| B. Readability `RenderHTML` -> structural representation | **Primary experimental path** | Parse selected HTML as a DOM and preserve heading hierarchy, prose, inline links/emphasis, table/list boundaries, images and explicit figure/caption/group relationships in document order |
| C. Original source DOM -> structural representation | Diagnostic comparator / potential selective fallback | Check which meaningful source relationships are omitted or altered by Readability; do not independently select a different article by proximity or URL heuristics |

**First question to answer:** Does the existing Readability-selected DOM still preserve enough structure to avoid the current `RenderText` -> paragraph block flattening loss? Inspect `RenderHTML` and `RenderText` side by side on identical input, using `cmd/readabilityprobe` and synthetic fixtures before writing a second selector. Do not assume Readability is always more faithful or that the original source DOM is always unambiguous.

### Structural conversion invariants

1. Parse and transform a **tree**, not sequences of flattened text and inferred-nearby captions. Retain order and parent/child relationships. Preserve explicitly evidenced `figure`/`figcaption`, caption/credit grouping, headings, time/update boundaries, inline formatting and links, lists and tables. Do not invent an association where source or selected markup is ambiguous.
2. Build a **nonrendering intermediate representation** before changing `internal/article` or `internal/ui`. Represent media ownership and groups explicitly rather than shoehorning them into anonymous adjacent paragraphs. Decide which model extensions are justified from multi-publisher examples.
3. Apply an allowlisted, security-preserving transformation to semantic objects, **not** a pass-through of publisher HTML into ScraperGo. Honor existing fetcher/redirect/access-policy and image-delivery boundaries; reject scriptable URLs and active HTML, preserve the unchanged production fallbacks.
4. Compare A/B/C on **identical saved source** and include positive and negative fixtures for ordinary prose, nested formatting/links, tables/lists, multi-image groups with and without unambiguous ownership, responsive alternatives, intentional repeated text, sidebars, live updates, malformed markup and denied/partial pages.
5. Score completeness, ordering, identity, attribution, excluded editorial material, safety, cost and rendering separately. A lower duplicate count or a higher image count is not by itself an improvement.
6. Only after demonstrated generality: propose a separately owner-approved production migration, browser/Print/PDF testing and fallback gates. No production switch or merge from this research PR without owner authorization.

### Next concrete development increment

Create an isolated offline comparator that feeds a **single captured HTML input** to the existing Readability parser, obtains `RenderHTML` and `RenderText`, and records a structurally typed representation of selected HTML without requiring the reader or network. Start with headings, paragraphs, inline links, `figure` and `figcaption`, lists and tables; explicitly flag unsupported media relationships. Compare this to unchanged control output and original-source-DOM evidence. Avoid evolving the previous wrapper proximity logic into the new converter.

This is a pivot in the existing [Issue #125](https://github.com/richard-mauri/scrapergo/issues/125) / draft [PR #126](https://github.com/richard-mauri/scrapergo/pull/126), not a request to close Issue #124 or reuse its defect branch. The owner-private Ars/New Scientist captures must remain on the Mac; existing synthetic fixtures may be committed. Existing prototype results are historical evidence and should not be cited as proof of a structural extractor's correctness.

---

## Matched A/B/C image identity and order audit (new offline experiment)

The previous owner Ars M6 replay found **20 image nodes** in Readability selected HTML, one explicit `figure`/`figcaption` pair, 66 paragraphs, 5 section headings, 28 links and 2 tables. This proves structural nodes survive `RenderHTML`, not whether the 20 images are the same images in source or reach the final reader correctly.

The offline `cmd/selectedstructure` now uses ScraperGo's actual saved-source preprocessing, **`structured.StripEmbeddedMedia`**, for every branch, and reports two side-by-side **exact resolved URL identity and occurrence-order** comparisons using the same local HTML. Source (C) requires exactly one explicit `article` scope; if there are zero/multiple `article` elements, `sourceStatus=inconclusive` instead of guessing a candidate. Selected (B) walks the typed semantic tree returned from the same Readability `RenderHTML` parse. The actual production control (A) runs **`bootstrap.ExtractWithEvidence` and `reader.DefaultProcessor().Process`** offline on the same preprocessed input; the comparison uses its final `article.Article.Blocks` with `Kind=="img"`, and reports the standalone `Article.ImageURL` as a **separate boolean**, because a lead image is not necessarily a distinct editorial block.

New JSON sections are `sourceSelectedImageAlignment` and `baselineImageAlignment`. Each reports counts of exact resolved URL occurrence matches, missing/extra images, `orderPreserved`, and source/selected index pairs **without ever printing source URLs, image URLs or publisher text**. Same URL occurring multiple times is paired one-to-one in order; different URL variants are *not* declared the same image. Explicit figure and figcaption counts are shown only when the node tree itself establishes ownership. The canonical comparison cannot infer a caption from a neighboring paragraph. The source comparison tests article scope and filters excluded active/hidden regions, so these figures are **not an unfiltered inventory of every `img` on the page**.

This comparison is evidence of **exact URL presence and sequence**, not pixel identity, visual duplication, correct CSS visibility, real caption attribution, editorial completeness, or browser correctness. In particular, an image represented solely by canonical `Article.ImageURL` can be absent from `img` blocks; that difference is intentionally recorded instead of hiding it. Source and selected parsing are separate observations; neither substitutes for a manual source-backed review.

The new tests cover exact URL matches in order, reordered images, omitted/new images, repeated occurrences, query-string variants, unsafe URLs, missing/ambiguous source article scopes, separate canonical lead image and duplicate canonical image blocks. These are synthetic only; owner-side matched Ars and unrelated publisher results remain **not run**.

To run locally without fetching publishers, after pulling this branch:

```bash
go test -count=1 ./...
go run ./cmd/selectedstructure \
  -input /tmp/pr120-ars-m6.html \
  -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/'
```

Review only the numeric/index JSON; do not share the private raw publisher capture. Full Go CI confirms tests/build; Safari reading, image ownership and Print/PDF are still owner-only future acceptance. The production code and default extractor remain unchanged.

---

## Initial implementation of the selected-HTML structural converter (Phase 2)

The owner-approved pivot has now produced an **isolated offline-only converter** at `internal/sourcepreserve/selected_html.go`. Its input is specifically **Readability `RenderHTML` output** (not raw source article selection); `ConvertSelectedHTML` parses with the existing `golang.org/x/net/html` parser and builds a typed tree of semantic nodes, keeping child order and explicit parent ownership. It observes headings, paragraphs, inline links and emphasis, lists/list items, tables/rows/cells, time elements with `datetime`, and explicitly nested `figure`/`img`/`figcaption`. Unknown wrappers are transparent, so it must not pretend to retain CSS layouts. It retains a bare `img` as an image with no invented caption. It excludes executable/interactive/noneditorial nodes under the existing experimental exclusion rules, checks link/media HTTP(S) URLs with `sourcehtml.ResolveURL`, and returns **data** rather than browser-renderable HTML.

The new `cmd/selectedstructure` runs the existing Readability dependency on an **already saved local HTML snapshot**, obtains `RenderHTML` **and** `RenderText` on the same parsing result, runs the selected-HTML converter, and prints **aggregate node-kind counts and flat-text character count only**. It deliberately does not print article text, image URLs, or a full copy of the semantic tree. Run from the repository root:

```bash
go run ./cmd/selectedstructure \
  -input /tmp/pr120-ars-m6.html \
  -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/'
```

The summary may be shared after reviewing it; never upload the owner's private publisher capture. Deterministic tests check explicit two-figure caption ownership without proximity matching, inline formatted links, repeated editorial paragraphs, heading hierarchy, time, ordered lists, tables, unsafe `javascript:`/`data:` URLs, active markup removal, and ambiguity of an image adjacent to an editorial paragraph. Neither the converter nor the command is called from `internal/extraction`, `internal/reader` or `internal/ui`. No changes to the production canonical `internal/article` representation have been proposed.

**Limitations / next decision:** This is a small proof of **DOM semantics retention**, not yet a complete article extractor, robust CSS/HTML sanitizer, real photo-caption pairing fix, full list/table semantics validator, complete inline-whitespace normalizer, or side-by-side fidelity evaluation. Readability may still remove source structures; compare the selected-HTML output against source structural inventories and the current production final canonical blocks on identical HTML. Include negative/ambiguous inputs across multiple publishers before proposing any migration. The old ancestor-count diagnostics remain frozen research evidence and are **not** the converter.

---
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

## Historical proposed source-DOM experiment — superseded by pivot above

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

## Historical offline prototype (diagnostics only; frozen by pivot)

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

## Third owner Ars replay and bounded gallery diagnostic

The owner-local unchanged Ars M6 HTML yielded `ambiguousFigures=19`, `unwrappedImages=19`, **`compactImageTextWrappers=6`**, **`otherUnwrappedImages=13`**, `skippedNoneditorial=28`. The six compact wrapper cases show a small ancestor with exactly one image and an adjacent text-bearing child. This does **not** prove six captions, six unique photos, or six safe-to-recover figure structures. The other thirteen may be in galleries, responsive alternate branches, or genuinely independent images.

The offline-only report now further partitions `otherUnwrappedImages` into `otherInMultiImageGroups` and `otherWithoutBoundedGroup`. Its bounded group test requires 2–8 image descendants, separate image-bearing direct children, and a maximum of four image-to-container edges, stopping at article/main/body. It does **not** preserve or attribute a caption, deduplicate image identities, inspect CSS, or change extracted `Items`.

**Important negative finding in synthetic CI:** A section with an unrelated editorial paragraph and one photograph can satisfy the weak `compactImageTextWrappers` rule. The diagnostic must never use that counter as authorization for caption recovery. The test now explicitly expects weak compact evidence while verifying that two photos in separate article regions are **not** grouped as one gallery. This prevents the experiment from silently treating text proximity as source-proof of captions. The CI regression was in an overstrong test expectation, not the source classification behavior.

The next owner replay only needs the numeric summary, including `otherInMultiImageGroups` and `otherWithoutBoundedGroup`, to choose the next source-backed structural investigation. No production change or merge is authorized.

## Fourth owner Ars replay: all noncompact images are in bounded groups (Nov 11)

The owner reran the unchanged private Ars M6 snapshot and reported `ambiguousFigures=19`, `unwrappedImages=19`, `compactImageTextWrappers=6`, `otherUnwrappedImages=13`, **`otherInMultiImageGroups=13`**, `skippedNoneditorial=28`. Thus none of the 19 image encounters lack structural context according to the deliberately weak diagnostics: six have a nearby single-image/text sibling, and all 13 remaining ones appear within bounded 2–8-image containers. **That does not prove any are unique editorial photographs, that associated text is a true caption, or that the 28 skipped subtrees are correctly classified.** The weaker compact wrapper classifier has already produced a synthetic false-positive with unrelated article prose.

### Next offline-only audit: exact image URL identity

The inspector now reports `resolvableUnwrappedImages`, `uniqueUnwrappedImageURLs`, `repeatedUnwrappedImageOccurrences`, `unresolvableUnwrappedImages`, and `boundedMultiImageGroups` as aggregate numeric counters (no new publisher text or image URLs added to the report). Unwrapped image identity is based on **exact resolved HTTP(S) `img src` URL strings**, using ScraperGo's existing `sourcehtml.ResolveURL`; this is not visual identity or proof of responsive duplicates, and it does not interpret `srcset`, `picture`, CSS backgrounds, query-size variants or bytes. Repeats are extra occurrences beyond the first matching exact URL; unique plus repeated occurrences equals resolvable unwrapped occurrences. The bounded-group count is the number of distinct nearest eligible image-bearing container nodes observed **for the noncompact branch only**, not a count of independently validated editorial galleries. It does not reflect excluded HTML or `figure` items. Synthetic tests cover literal duplicate URLs, distinct photographs with the same caption, query-string variants, unsafe URLs, source order and preserved prose.

**Next owner step:** pull the same draft PR branch and run the same offline `structureprobe` with `| tail -n 22`, then return **only summary counter lines**. Compare repeated-image occurrence count and bounded-group count to the already established 6/13 structure; do not export the full publisher-text JSON or assume exact-URL equality is enough to collapse representations. No production extractor, reader or UI changes.

## Latest owner image identity evidence and content-free structural trace

Same owner-private Ars M6 snapshot reported **19 unwrapped image elements with 19 unique exact resolved `img src` URLs** (zero exact URL repeats), **6 compact image/text sibling candidates**, **13 images in bounded multi-image groups**, and **7 distinct nearest bounded groups** among the noncompact images. `skippedNoneditorial=28`. These are *structural counts*, not 19 distinct visual photographs or seven verified editorial galleries. Different URLs may serve the same photograph/crop/size; no browser CSS or `srcset` equivalence has been established.

A new `cmd/structureprobe -structure-trace` **opt-in offline mode** emits only a report with per-run numeric image ordinals, the direct parent element's tag name, structural pattern (`compact-text-sibling`, `bounded-multi-image-group`, or `unwrapped`), numeric bounded-group identity, and group counts of images, image-bearing branches and text-bearing branches. The trace **does not emit publisher article text, captions, credits, class/id values, source URLs or image URLs**; it does not fetch data or modify the article. The numeric group IDs are assigned in first-observed traversal order; they are not persistent identifiers or proofs of gallery identity. The same weak compact-text heuristic can match unrelated article prose, so this report is **diagnostic only** and must not assign captions or remove duplicate images.

Local command after pulling the branch (prints a small content-free report and does not write any files):

```bash
go run ./cmd/structureprobe -structure-trace \
  -input /tmp/pr120-ars-m6.html \
  -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/'
```

You can share this structural JSON with no article text or links. It is still a local diagnostic; review output before sharing for privacy, and do not upload the private HTML. New synthetic tests assert bounded group identity/shape, per-photo ordering, omission of source content/URLs, refusal of ambiguous article selection/invalid base URLs and exclusion of hidden/advertising and already-recognized figure elements. This is the first trace capable of showing *how images are structurally grouped* rather than merely counting encounters. No production path or experimental `Items` changed.

**Next:** use this trace to determine if actual Ars photo pairs, credits and captions are represented in the same DOM branches or separated by responsive/gallery wrappers; corroborate against original Safari publisher view before attributing any caption. A later prototype must still compare editorial completeness and safety against merged Readability on identical inputs.
