# Offline single-article duplicate provenance — Issue #119 / draft PR #120

Run this diagnostic **only against previously captured, authorized publisher HTML**. It invokes the established ScraperGo extraction and reader code without fetching a URL, without contacting any remote publisher, and without changing server behavior. The diagnostic prints final candidate block indices/types, a short nonreversible text digest, and the recovery stage counts. It does not print article content or persist raw HTML. It is not a production deduplication fix.

First, from the existing PR #120 checkout, capture each public article with the existing security-validated command. Use **distinct output paths**, as `htmlcapture` will not overwrite files:

```sh
cd ~/projects/scrapergo
go run ./cmd/htmlcapture \
  -url 'https://www.newscientist.com/article/2598048-a-true-paleo-diet-would-include-a-lot-more-insects/' \
  -out /tmp/pr120-paleo.html
go run ./cmd/readabilityprobe \
  -input /tmp/pr120-paleo.html \
  -url 'https://www.newscientist.com/article/2598048-a-true-paleo-diet-would-include-a-lot-more-insects/' \
  -out /tmp/pr120-paleo-readability
go run ./cmd/duplicateprobe \
  -input /tmp/pr120-paleo.html \
  -url 'https://www.newscientist.com/article/2598048-a-true-paleo-diet-would-include-a-lot-more-insects/'
```

Then independently inspect the original Ars example:

```sh
go run ./cmd/htmlcapture \
  -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/' \
  -out /tmp/pr120-ars-m6.html
go run ./cmd/readabilityprobe \
  -input /tmp/pr120-ars-m6.html \
  -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/' \
  -out /tmp/pr120-ars-m6-readability
go run ./cmd/duplicateprobe \
  -input /tmp/pr120-ars-m6.html \
  -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/'
```

If source fetching is rejected or returns incomplete challenge HTML, **stop and report it**; do not bypass access restrictions. `htmlcapture` may return source HTML different from a browser-rendered article; the probe does not emulate network-dependent PDF fallback or full HTTP page gating. Distinguish these offline outcomes from the full `/api/read` JSON.

You can paste just `duplicateprobe` stdout and `readabilityprobe`'s metadata `report.json` (not the HTML) to the PR reviewer. Private captures may contain copyrighted full-page text and must not be uploaded or committed indiscriminately. The block-index findings are a reproducible aid for inspecting the source locally, not a safe reason to remove text without checking legitimate occurrences.

If source snapshots differ from the earlier 100-URL run, record the discrepancy and retest with the same snapshot. Keep all source-backed fixes on this single draft PR #120.

## Ars M6: compare source and Readability HTML without printing article text

After a secure capture, run against the **same** saved HTML; do not recapture:

```sh
go run ./cmd/duplicateprobe \
  -input /tmp/pr120-ars-m6.html \
  -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/' \
  -source-evidence
```

The optional `-source-evidence` report prints, for each pair of final duplicate-candidate blocks, how many source HTML and Readability-selected HTML **paragraph** or **figcaption** elements contain the shorter matched text. It uses the current private source HTML and the same Go Readability library but does not expose article text or write output files. Because this is a conservative substring presence test, zero counts can occur when HTML text serialization/inline formatting differs. Counts can also include repeated occurrences that are editorially legitimate. A single paragraph plus a caption in the original source is not the same problem as two independent editorial paragraphs. **Do not treat these counts as a deletion instruction.** Inspect raw source locally if the report is ambiguous.

### Source text that is not inside a paragraph

The `-source-evidence` report also displays **Source minimal tags** and **Readability minimal tags**. These counts search for the shortest matching source element across *all nonblocked tag types*, rather than only `p`/`figcaption`. The text is concatenated across inline HTML nodes without injecting spaces between tags, and parent containers are not double-counted when a child already contains the full match. For example `div:2` means two independently matched smallest `div` subtrees; it does **not** prove those copies are editorially identical in meaning. Script/style/navigation text is excluded under existing source parser blocked-tag policy.

If the source has no matching elements even with this broader scan, the unresolved explanations include text/entity normalization differences, publisher generated content, or Readability's own synthesis. Do not infer the cause from an unmatched count alone. Never upload or commit the full captured HTML without permission.

## Stage 1/3: element ancestry, ordering and representation provenance

On the owner's Mac, **reuse the existing private HTML capture**. Pull the current PR branch, then run:

```sh
cd ~/projects/scrapergo
go run ./cmd/duplicateprobe \
  -input /tmp/pr120-ars-m6.html \
  -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/' \
  -source-provenance
```

The new `-source-provenance` flag compares **each member** of a duplicate-candidate pair (not just the common shorter text). It prints the earliest/minimal matching HTML elements from the original saved source and the separately rendered Readability HTML, including a local document-order ordinal (`#N`), index-qualified element path (`root/html[1]/body[1]/article[1]/...`), immediate parent tag, nearest article/section/figure/figcaption/etc. container tag, origin (`text`, `attr:alt`, `attr:title`, `attr:aria-label`, `attr:data-caption`), and match mode (`text` or `compact`). The compact mode ignores whitespace across inline nodes; it is weaker and should not alone justify deleting text.

The output contains **no text samples, publisher URLs, class/id or attribute values**; at most eight locations per passage are printed, and truncation is explicit. The command makes no network calls and writes no files. Paths and ordinals are **independent within source versus selected HTML**; a matching index on each side is not the same DOM node. The parser is ScraperGo's legacy source HTML tree, which can misinterpret malformed publisher markup, and Readability provides no stable source-node ID mapping. Therefore the probe can establish whether source contains one or multiple text representations, but **cannot conclusively prove** that a particular selected paragraph was derived from one exact source node. A zero-match case is still inconclusive; text may be supplied via alternate attributes or transformed with punctuation changes.

Interpret the results conservatively: two independent source paths might be intentional editorial repetition, caption+alt, hidden alternatives, or unrelated page furniture. If one source passage appears in two selected paragraphs, Readability duplication is a plausible hypothesis but requires inspecting the precise source/selected containers. Do not expose copyrighted captures or implement a generic string-dedup rule. Stage 2 (production correction) and stage 3 (full regression/Safari validation) remain **out of scope for this probe change**.

### Stage 2 evidence gate: media adjacency and visibility hints

The same `-source-provenance` command now prints two extra safe fields for each source and Readability match: `media=parent-img:N,figure:N,figcaption:N,siblings:N` and `visibility=...`. These describe the immediate parent subtree and explicitly declared HTML visibility (the `hidden` attribute, `aria-hidden=true`, or simple inline `display:none` / `visibility:hidden`). A value of `no-explicit-hint` **does not mean browser-visible**: responsive stylesheets, JavaScript and gallery state cannot be reliably inferred from captured HTML. Image URLs, CSS classes, publisher text and attribute values are never printed.

The current Ars source selected text comes from `div`/`span` layouts, whereas `removeFlattenedFigureCaptions` only reconciles explicit `figure`/`figcaption` descendants in Readability-selected HTML. This broader diagnostic is needed to establish a general structural predicate before changing production code. **Do not automatically drop repeated `p` blocks or photographer credits from source text**; identical captions may belong to different images, and publisher text repetition can be intentional.

**Legacy parser limitation:** this diagnostic currently recognizes `hidden` only if parsed as an explicit name/value attribute (for example, `hidden="hidden"`). A bare Boolean `hidden` attribute may be lost by `sourcehtml.Parse`, and stylesheets/media queries and dynamic visibility are not evaluated. Hence `no-explicit-hint` always means **visibility unknown**, not verified visible.

### Photo ownership scope — same original capture

With `-source-provenance` the offline report also displays `publisher photo scopes` for each matching block. A `source image#N` is an ordinal within **that source HTML snapshot only**, not an image URL, hash or canonical media identity. It identifies the nearest ancestor subtree containing exactly one `img`, reports its tag, path and ancestor-edge distance. When no single-image subtree is found, the probe reports a nearby multi-image group or `no image ancestor before article boundary`.

This differentiates text adjacent to one image, alternative caption text sharing a parent image subtree, multiple photographs with identical captions, and genuinely independent article paragraphs. **Matching photo ordinals across two minimal text matches are correlation, not proof that both are captions**. Where a source ancestor covers a multi-image gallery, the owner remains unknown; CSS visibility and caption credit ownership are not reconstructed. The source parser can misinterpret malformed HTML, so verify any proposed rule against the preserved local HTML and published Safari presentation.

The probe has regression cases for one photograph with alternative caption markup; two distinct photographs with identical captions; unrelated repeated editorial prose despite an article image; a two-photo group with ambiguous ownership; and exclusion of image URL, class, id and raw content from the report.

Run the **same command** as above on the unchanged `/tmp/pr120-ars-m6.html`; no new publisher request or credentials are required. Do not begin text removal solely because two nodes share a photo ordinal. A general-purpose extractor correction also requires source-supported image-caption-credit relationships and negative content-preservation tests.

## Stage 2 candidate: source-backed caption/credit reconciliation

The current draft PR #120 now contains a **provisional production extraction fix** in `internal/extraction/figure_caption_credit.go`, invoked at the existing `caption-reconciliation` checkpoint in `internal/extraction/readability_pipeline.go`. It targets two source-backed patterns: `caption → credit → caption+credit` and two successive combined caption-credit representations that differ only in whitespace. The first keeps caption and credit separately; the second keeps the better-spaced representation.

The filter activates **only** when the original publisher HTML *and* Readability-selected HTML each contain an image-associated caption/credit representation and a separate alternative in a compact one-image scope. The original source's alternative cannot be an independent `p` paragraph, and any ambiguity across multiple matching images disables removal. It does not rely on article URLs, classes, photographer names, or global text deduplication. Image blocks, distinct pictures sharing the same caption, unrelated repeated prose and linked/structured blocks are protected by regression cases. The exact Ars M6 block 4/5 pair remains unresolved independently.

**Do not infer that Ars has been fixed from synthetic CI tests.** The saved original publisher snapshot has not yet been run through this new production candidate. On the owner Mac, update the existing `fix/119-caption-duplication` branch, run `go test -count=1 ./...`, then run:

```sh
go run ./cmd/duplicateprobe \
  -input /tmp/pr120-ars-m6.html \
  -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/' \
  -source-provenance
```

The previous same-snapshot baseline was **68 canonical blocks, 1 exact and 3 overlapping candidate pairs** (4/5, 17/19, 22/24, 25/27). Compare the new `caption-reconciliation` transition, final pair indices and content presence. A count decrease alone is not sufficient: inspect **both** distinct photo captions/credits, left/right association, images, and real story paragraphs in Safari and Print/PDF. Do not merge or deploy until the owner verifies the output and reviews the separate exact 4/5 pair. New Scientist same-snapshot regression and the 100-URL batch are part of Stage 3, not replaced by synthetic tests.


## Stage 2 failed Ars candidate — inspect exact block eligibility

The owner re-ran the same saved Ars HTML **after the provisional Stage 2 extraction correction**. The result remained **68 blocks, 1 exact pair (4/5), 3 overlapping pairs (17/19, 22/24, 25/27)**, with `caption-reconciliation` reporting **no changes**. The candidate has **not demonstrated improvement on the triggering publisher**, despite synthetic Go CI passing. Do not begin Stage 3 or merge based on the synthetic result.

The `-source-provenance` mode now prints a separate **Candidate shape evidence** section **before** DOM ancestry. For each flagged pair it reports the number of intervening blocks, the middle kind when exactly one block intervenes, whether `caption + middle` reproduces the later text exactly or after whitespace compaction, and whether adjacent/other pair texts are equal exactly or only with whitespace removed. It does **not** print publisher text, evaluate source/selected image ownership, authorize deletion, or alter extraction.

A triple mismatch, nonconsecutive pair, or compact-only match narrows why the provisional correction did not engage. If all shape tests match, the remaining suspect is the source/selected HTML guard; retain its safety requirement and inspect the corresponding image-caption subtree before editing production behavior. The exact 4/5 pair is still unsupported by minimal source matches. Re-run the same private capture offline, then compare the shapes and source photo scopes and record exact evidence. For any actual fix, require complete source-backed article-output regression tests and owner Safari Print/PDF acceptance.

### Actual production guard verdicts — after shape confirmation

Owner Mac's unchanged Ars capture (68 blocks, 1 exact and 3 overlapping pairs) has now **confirmed that all three photo caption pairs are EXACT three-block patterns**: 17/19, 22/24 and 25/27 have one middle `p`, and the first caption plus middle credit exactly equals the later combined paragraph. This rules out the triple-shape check as the blocker. The reconciliation stage still changes no blocks. The unrelated p4/5 exact pair has gap zero and zero source minimal matches.

The `-source-provenance` report now prints `production triple gates:` immediately under each applicable overlap pair, using `extraction.InspectCaptionCreditTriple` with **the same source and selected HTML scope predicates as production**. `shape` checks the full block-level predicate, including rich combined content; `rich-combined` reports whether parts/links prevent removal. `source-image-scope` and `selected-image-scope` show which of the two strictly required structural checks passes; `eligible` is their conjunction with `shape`. The report has no prose, source URL, HTML or image asset identifiers. It is observational only and makes no changes to the extraction pipeline.

If a scope is false, do not simply weaken the gate: investigate whether source/selected image and separate caption branches are nested differently, whether the two representations are actually the same photo, and whether allowing that structure would preserve real editor prose, separate-photo captions and original left/right association. This pass does not validate the real Ars fix or resolve p4/5.
