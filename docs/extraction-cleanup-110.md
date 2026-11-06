# Issue #110 — staged extraction cleanup audit and evaluation

**Status:** Stage 1 source audit complete for the primary Readability extraction path; Stage 2 narrowly scoped to provable unused state and a regression guard. Stage 3 is an evaluation, **not** a rewrite approval. See [Issue #110](https://github.com/richard-mauri/scrapergo/issues/110) and [PR #111](https://github.com/richard-mauri/scrapergo/pull/111). Baseline: merged PR #109 on `main` (`6cd6229`). The CBS screenshot demonstrated separated time, prominent heading, and intact linked prose; do not change those semantics.

## Stage 1: reachable-path inventory

This is a **static source-path inspection**, not an exhaustive dynamic call graph or proof of publisher compatibility. It identifies the entry-point branches and distinct responsibilities that prevent deleting apparently overlapping recovery functions.

| Production call / module | When it runs or is selected | Why it is still live |
| --- | --- | --- |
| `ExtractWithReadability` | Main readability extraction entry; rejects malformed URL; attempts `RenderText`; falls back if parsing, output or word-count gate fails | Controls the accepted extraction/fallback contract |
| `RecoverStructuredDocument` | Before the other recoveries; may replace the selection when repeated editorial headings and source corroboration warrant it | Rebuilds a dense multi-section editorial subtree; `structuredRecovered` skips other stages |
| `RecoverLeadingParagraphs` | Only if structured recovery was not chosen, before the 60-word gate | Restores source-backed introduction before Readability's first retained paragraph |
| `RestoreSourceTables` | On `count >= 60`, independent of structured-recovery flag | Converts corroborated flattened table cells into row/column blocks |
| `PreserveSelectedHTMLLinks` | If `RenderHTML` succeeds inside the accepted Readability branch | Retains inline link targets from selected markup |
| `PositionArticleFigures` | With selected HTML, only when structured recovery was not chosen | Positions images relative to retained source text |
| `SupplementGalleryFigures` | Only within above figure path when URL path includes `/gallery/` | Recovers additional captioned gallery figures; existing narrowly scoped behavior, not a new rule |
| `RecoverEditorialContinuity` | Accepted Readability result when no structured-document replacement was chosen | Recovers source-backed missing article paragraphs and images, otherwise anchored tail |
| `RecoverInstitutionalTail` | After continuity, regardless of structured-document choice | Handles adjacent, source-backed institutional explanation |
| `RestoreHeadingAnchoredProse` | After institutional tail, before the new heading/time overlay | Can add omitted paragraphs only when **three already-typed** heading anchors are present |
| `PreserveReadabilityStructure` | At end of accepted extraction if selected HTML was rendered | Maps uniquely corroborated `h2/h3/h4` text back to headings; separates a matching `time` suffix or standalone timestamp |
| `LegacyExtract` | When Readability does not return a qualifying article | Preserves fallback for unsupported markup |
| `PartialHTMLArticle` | When both Readability and legacy fallback fail | Constrained public excerpt after metadata and canonical-URL corroboration |
| `internal/reader/live_updates.go` | Separate post-extraction reader stage; existing `/live/`-gated long-duplicate cleanup | Does not infer headings/timestamps; must retain source-supported updates, untouched in this issue |

The Readability branch calls `RenderText` first and `RenderHTML` later. A successful `RecoverStructuredDocument` does **not** prevent the final `PreserveReadabilityStructure` call. The pipeline reparses already-fetched HTML in multiple recovery stages; this is real redundancy in work, **not** proof that those stages are dead. The previous heading-anchored prose recovery runs before the new heading overlay: recovering a heading later cannot make that earlier stage execute again.

## Classification and disposition

| Candidate | Classification | Disposition |
| --- | --- | --- |
| Local `evidence.text` in `PreserveReadabilityStructure` | **Provably unused state**: written in evidence literal, never read in the function | Removed in Stage 2; no effect on heading/time selection |
| `headingCounts`, `blockCounts`, `recoverable` in the structure overlay | **Live validation state** | Retain: unique matching prevents ambiguous heading promotion |
| `splitTrailingSourceTime` and its inline-parts preservation | **Live, conditional** | Retain: part boundaries and links are more fragile than an unused field; future simplification requires equivalence evidence |
| `RestoreHeadingAnchoredProse` and final heading restoration | **Potentially overlapping but live** | Retain; investigate ordering under Stage 3 instead of deleting |
| Separate source DOM parses across recovery stages | **Redundant computation with live distinct functions** | Defer shared parsing/cache design until profiling and explicit owner approval |
| Structured-document, continuity, leading, institutional, figures, tables, links, legacy, and partial recovery | **Necessary fallback/branch behavior under current implementation** | No removal without corpus, caller, and regression evidence |
| `RenderText` followed by heading/time repair from `RenderHTML` | **Architectural duplication, not dead code** | Compare with direct source-backed HTML-to-block construction only in Stage 3 |

This list does not claim that no other unused function exists anywhere in the repository. A complete whole-repository dead-code report would require a reliable call-graph/build-tag analysis; this audit focuses on Issue #110's active extraction path.

## Stage 2: minimal mechanical change and behavior guard

Removed only the unused evidence struct field and literal initializer in `internal/extraction/readability_structure.go`. Added `TestPreserveReadabilityStructureDoesNotMutateInput`: the output separates a verified timestamp, promotes the verified heading and preserves its inline link, while the caller's original blocks and parts remain unchanged. Existing tests exercise tanker-like boundaries, independent timestamps, linked paragraphs, ambiguous/duplicate headings, normal articles, and a full extraction path. These are synthetic fixtures, not real publisher acceptance.

No older recovery functions, fallback stages, URL conditions, dependency boundaries, or production UI behavior were changed. Any future cleanup needs a measured comparison and distinct owner authorization.

## Stage 3: architectural evaluation only

**Option A — Retain the text-first extractor and structural overlay (current).** Low disruption; established fallback behavior and evidence matching stay in place. Cost: at least two Readability representations (`RenderText` and `RenderHTML`), text matching back to HTML headings, and multiple downstream DOM traversals. It is robust for the demonstrated CBS update but can miss malformed, repeated, or absent source heading/time semantics.

**Option B — Construct canonical article blocks in DOM order from Readability-selected HTML.** May remove the *need* to flatten then restore heading/time structure and could reduce repeated text matching. However, it must preserve existing editorial selection, link parts, tables, figure positions/captions, galleries, text continuity, word counts, partial-content/access gates and original fallback. Reading-order and noisy-node filtering are nontrivial. Do not assume older modules become dead; design an explicit equivalence matrix and profile first.

**Option C — Share parsed selected DOM across existing recoveries without changing output.** Potentially smaller intermediate refactor than B, but requires API redesign and verification of mutable DOM lifetime/ownership. Do not undertake under the low-risk cleanup authorization.

**Complexity evidence:** The source-backed overlay is about 100 lines; the main Readability extraction function is about 110 lines, and additional recovery modules each implement different conditional functions. Replacing the overlay alone is therefore **not** demonstrably a net complexity reduction. No runtime benchmarks or per-publisher output-equivalence measurements were performed.

**Suggested order if future work is approved:** capture safe golden article-block results across a corpus; cover (1) ordinary and multi-update articles with different HTML layouts, (2) repeated/ambiguous headings and missing time tags, (3) inline links, media roles, figures, galleries, table cells, (4) incomplete/blocked pages and fallback, (5) live duplicate handling, and (6) Safari reader and print/PDF. Compare candidate outputs exactly, include resource/performance measurements, and keep reversible rollback. Discuss any removal or reordering with the owner before implementing.

## Validation boundaries

Go CI must pass on the exact latest PR head. Previous PR #109 passed CI and was visually accepted for the CBS example; that is baseline evidence, not proof of the new cleanup. The current regression tests run locally in CI. This issue did **not** retrieve a fresh publisher page; the `cmd/htmlcapture` tool is opt-in and source captures must remain private and permitted. Mac/Safari live-site/print acceptance is performed by the owner, not GitHub Actions.
