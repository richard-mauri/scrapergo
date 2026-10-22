# Architectural migration plan (v0.2.85 baseline)

## Current architecture

The main executable is a single Go package that shares `Article`, `Block`, `TextPart`, and `InlineLink` across fetch, extraction, image reconciliation, classification, diagnostics, API, and rendering. Multiple sequential transformations mutate `Article.Blocks`, and order is meaningful. Until captured live-page fixtures exist, a broad move or consolidation risks subtle behavioral changes.

## First boundary: canonical article document

`internal/article` now owns the transport-independent types `Article`, `Block`, `TextPart`, and `InlineLink`. Root compatibility aliases retain the same exported fields, existing JSON tags, function signatures and tests. The domain package imports no HTTP, parser, UI, or persistence packages. No extraction algorithm, image ranking, link recovery or runtime logic changes in this PR.

This is a **first real import boundary**, not a claim that the pipeline has already been decoupled. The root code is still mutually dependent through functions and mutable article blocks. Future migrations can import `internal/article` directly rather than `package main`.

## Ownership design for subsequent PRs

1. **Retrieval:** input URL, controlled network fetch, final URL, raw HTML, and access evidence. No editing of canonical articles.
2. **Extraction:** readable text and structured source evidence. No HTTP handlers.
3. **Media:** candidate images and provenance, source identity, image roles. It should not independently reorder canonical blocks after final assembly.
4. **Document assembly:** one responsible stage that reconciles body, links, contributor data, and media and then emits an article value.
5. **Presentation/API:** serialize/render the assembled document without reinterpreting extraction evidence.

Before migrating the extraction pipeline, capture representative real-page fixtures and expected images/bylines (including CNN and NPR) with sensitive values removed. Verify all existing tests, Go build, binary build, local Safari behavior and Cloudflare tunnel. Make each step a separate PR; do not change publishers' access controls or write publisher-specific extraction logic.

## Current constraints

- Preserve `module clean-article-reader` until a separately reviewed rename is justified.
- Preserve root `main.go`, `go build .`, `./devbuild.sh`, and `./scrapergo.sh` while the migration proceeds.
- No version bump for internal nonbehavioral changes.
- Never merge before CI and owner review.

## Phase 2: initial extraction and source evidence handoff

The `internal/extraction.Result` type now separates `Article` from `SourceEvidence` (`SourceURL`, `MetadataImageURL`). Production rendering and the image-trace diagnostic both consume the same handoff function, `extractWithEvidence`, rather than recomputing metadata within their own normalization sequence. Source metadata remains an input to existing image normalization, not a forced lead image. Existing downstream reconciliation still mutates `Article.Blocks`; consolidating those mutations into a document assembler is the next phase, after frozen fixtures validate their observable outcomes. This change does not alter fetch policy, publisher handling, image rules, or the JSON API.

## Phase 3: one article-assembly orchestration point

`document_assembler.go` owns the v0.2.85 post-extraction reconciliation sequence used by the production API: photographic recovery, deduplication, link repair, portrait recovery, embedded-player filtering, image-role classification, byline reconciliation, source-leading figure recovery and final normalization. Its input is `internal/extraction.Result` plus the already-cleaned source HTML, and its output is a canonical `Article`. A characterization test compares the assembler output with the prior stage order and asserts the editorial lead image. No algorithm changes are intended. The diagnostic image trace intentionally retains its separately instrumented stages for debugging; converging its instrumentation with the assembler without losing intermediate observations is future work. Package decomposition, mutation ownership, and fixtures from real publishers remain outstanding.

## Phase 4: order-preserving transformation stage contract

`internal/assembly` defines a `Stage` interface, an adapter for legacy transformation functions, and a pipeline runner. The runner copies the initial article blocks and nested link/part slices before running transformations, protecting the extracted document from subsequent in-place edits. `document_assembler.go` enumerates the existing eleven passes as named stages in their previous order; the same Go implementations and image/source evidence are used. This makes stage orchestration and ownership testable without prematurely migrating their tightly coupled implementations into incompatible packages. Diagnostic stage instrumentation remains unchanged. Follow-up work should build frozen source fixtures and gradually migrate individual stages behind narrower domain interfaces.

## Phase 5: extract media identity and deduplication

`internal/media` now owns image delivery-URL identity and duplicate elimination as a dependency-light, independently tested package. The document assembler invokes `media.Deduplicate` directly, while existing root helpers delegate for compatibility with older callers and tests. This does not change which photos are selected or their order, and does not alter source-image role classification. The migration is intentionally limited to one low-coupling stage; further stages should move only after their dependencies and output invariants are characterized.

## Phase 6: contributor-header package

The contributor header recognition and contiguous preamble separation now live in `internal/contributors`. The assembler calls this independent, article-model-only package; root compatibility wrappers preserve old callers. This phase does not change byline policy, portrait classification, extraction behavior or rendering. Dedicated tests cover Parts/Links contributor anchors, editorial-image boundary, and preservation of article body. Wider contributor/image-role interactions remain separate and should be migrated only after representative fixtures are captured.

## Phase 7: parser-independent image-role evidence

Source image-role classification now has two responsibilities with a real package boundary: `source_image_roles.go` reads HTML nodes and exports ordered ancestor evidence (`media.ImageEvidence`), while `internal/media` decides author/editorial roles and applies them to the canonical article. Image source identity remains the existing implementation and is injected into the policy. The same ancestor priority, source order, editorial-over-author precedence, and preservation of existing lead roles are retained. This is a structural change; subsequent work can replace parser details independently and develop frozen real-page source fixtures before altering image heuristics. Regression tests cover role evidence and precedence.

## Phase 9: independent source HTML tree package

The source-tree parser, traversal, text normalization and metadata helpers have moved into `internal/sourcehtml`. Existing callers retain their `Node`, `parseHTML`, `find`, `textOf`, `meta`, and `normalize` interfaces as thin forwarding wrappers while the parser is now package-testable and reusable independently of the HTTP server. The application-specific noise and policy filters remain in the application package. The parser deliberately retains the legacy tokenization behavior; replacing it with a standards-compliant tree is a separate, fixture-gated decision. No extraction ranking or page classification changes are intended.

## Phase 10: parser-independent link restoration

`internal/links` now owns article text-key matching and restoration of missing anchors. The root `link_recovery.go` collects HTML evidence (selected paragraph text and resolved source anchors) using existing parsing and URL policies, then passes the observations to `links.Restore`. The policy package depends only on the canonical `article` model. It never fetches URLs, adds blocks, or overwrites existing structured links. Tests check exact-paragraph matching, duplicate suppression, and preservation of original links. Existing stage order, output JSON and launcher configuration are unchanged.

## Phase 11: independent editorial continuity policy

`internal/recovery` now owns the pure paragraph-order corroboration and conservative full-document replacement decision. The parser-dependent candidate enumeration, article scope checks, advertising filters, and anchored-tail recovery remain in `continuity_recovery.go` until they can be migrated with their source evidence. The root `orderedEditorialMatches` entry point delegates to the package for compatibility; no network fetches or publisher exceptions were added. The sourcehtml text normalizer is shared to keep equivalent comparison semantics. Unit tests cover ordered anchors, unchanged selections and rejection of truncated replacement text. This is an intermediate architecture boundary, not completion of the paragraph/continuity migration.

## Phase 12: structured text block conversion

`internal/structured` now owns conversion of a parsed HTML element to the canonical article block, including inline formatting, whitespace compaction, and links. URL resolution and the linkability decision are supplied as explicit functions by the legacy application adapter, preserving the existing security and link policy. This migrates a real part of structured-document reconstruction without moving its more coupled source-selection and image recovery algorithms. `semanticTextBlock` is a compatibility shim; downstream callers have unchanged signatures. Additional structured-document stages should migrate only after their policies and expected output are fixed in regression tests.
