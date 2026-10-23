> **Historical design record — not current instructions.** This document describes earlier migration/design states and may reference superseded branches, root package owners, aliases, and PR status. Current authority: [ARCHITECTURE.md](../ARCHITECTURE.md), [DEVELOPMENT.md](../DEVELOPMENT.md), and [HISTORY.md](../HISTORY.md).

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

## Phase 13: document media normalization ownership

`internal/documentmedia` now owns the article title, media deduplication, editorial lead precedence, and metadata rendition reconciliation pass. The existing image-source identity resolver remains injected from the application boundary, so delivery-specific identity and editorial ranking are not reimplemented. The root `normalizeArticleDocument` function forwards to the new package to preserve legacy callers and stage ordering. Package tests verify author portrait exclusion, title removal, editorial-first ranking, metadata rendition substitution and metadata-only fallback. This refactor changes no HTTP, scraper retrieval or version behavior; image recovery and source-specific evidence extraction remain separate for later review.

## Phase 14: injected HTTP presentation adapter

`internal/httpapi` now owns the image-metadata HTTP adapter and ordered image-candidate response. The application injects its existing article API handler, preserving access policy, extraction behavior, POST-only restriction, cache headers and error status/JSON behavior. The root imageMetadata and articleImageCandidates functions forward for compatibility. This is a first HTTP-boundary migration; article retrieval and primary /api/read handling still need independent service ownership. Dedicated tests verify the response contract, duplicate candidates, method restrictions and propagated errors.

## Phase 15: primary article-read HTTP adapter

The `/api/read` HTTP request and response contract now lives in `internal/httpapi.ReadHandler`. The executable injects the existing Google News resolver, secured downloader, publisher access-error classifier, hostname formatter and `writeExtractedArticle` assembly/response function. The adapter retains POST-only validation, the 4096-byte request limit, status/JSON error contracts, cache headers and request logging. No URL security checks, retrieval transport or content extraction policies are replicated in the HTTP package. `api` remains a small compatibility wrapper; primary article application-service ownership and wrapper pruning remain later work. Regression tests verify successful handoff, invalid methods/JSON and publisher-access error responses.

## Phase 16: centralized application handler wiring

`app_wiring.go` owns the production dependency setup for the article-read handler and the image-metadata adapter. The server now constructs its primary read handler once and passes that exact handler to the metadata route. The prior per-request constructor remains available only through compatibility functions used by existing tests/callers. Neither the URL resolver, protected downloader, publisher access policy, nor extraction writer changed. Router paths and legacy HTTP contracts remain intact. We intentionally retained compatibility wrappers referenced by tests and diagnostic paths; removing them without a comprehensive caller audit would introduce needless risk. This phase centralizes service wiring rather than pretending that all retrieval or parsing dependencies have moved out of main.

## Phase 17: stabilization review

The final architecture inventory and outstanding-debt list are maintained in [architecture-review.md](architecture-review.md). This milestone adds negative HTTP contract and assembly-isolation tests, documents the package dependency graph and release checklist, and deliberately makes no production implementation changes. Continued migration should be based on concrete remaining ownership and behavior requirements rather than a fixed PR count.

## Phase 18: remove root media/contributor compatibility shims

Deleted `image_dedupe.go` and `contributor_header.go`, which contained only root-package forwarding functions. Updated affected root regression tests and assembly-equivalence tests to invoke `internal/media.Deduplicate` and `internal/contributors.SeparateHeader` directly, matching production stage wiring. No new compatibility facade or callback injection was introduced. This is intentionally a narrow first deletion PR; source-HTML adapters that still perform parsing or transformation are not mere wrappers and must be migrated with their owning subsystem.

## Phase 19: direct structured media cleanup migration

Removed the root implementations of embedded media DOM stripping and serialized player-block pruning. `internal/structured` now owns `StripEmbeddedMedia` and `RemoveSerializedPlayerBlocks` along with the shared media-element recognition logic. Main extraction, diagnostic image tracing, assembly, and regression tests call these package functions directly; no compatibility wrappers or injected helper callbacks were created. The other source-dependent recovery algorithms remain root-owned pending separate cohesive migrations. Application version and network retrieval policies remain unchanged.

## Phase 20: delete image-metadata root compatibility entry points

The production image-delivery endpoint and root HTTP tests now invoke the application-wired `internal/httpapi.ImageMetadataHandler` directly via `newArticleMetadataHandler`, rather than a package-main `imageMetadata` forwarding function. Deleted `image_metadata.go`, including its redundant `articleImageCandidates` forwarder. The actual HTTP response and image-candidate selection remain owned by `internal/httpapi`; this change adds no replacement wrappers or duplicate image policy. Full migration of page selection and partial-excerpt extraction is deferred until shared URL and editorial policy can move with their actual owners.

## Phase 21: canonical document-link eligibility

Moved the complete `linkableDocument` URL-and-filetype eligibility policy from the root Readability implementation into `internal/links.IsDocumentURL`; removed the root implementation and updated root Readability and structured-text calls directly. Added package-local policy regression tests. No new compatibility wrapper or extra callback parameter was introduced. The rest of Readability/extraction, source image URLs and network validation are unchanged. Continued extraction migration must move cohesive dependencies rather than cloning root policy.

## Phase 22: link identity ownership and wrapper deletion

Moved the complete semantic paragraph identity implementation out of `semantic_links.go` and into `internal/links.SemanticKey`. Existing selected-HTML link reconciliation now calls that implementation directly. Removed the obsolete root `linkTextKey` forwarding function; `internal/links.TextKey` remains the canonical policy. Added package-local semantic matching tests. No compatibility adapters or changes to extraction URL and network policy.

## Phase 23: cohesive thumbnail processing ownership

Moved thumbnail dimension limits, image sniffing, bounded pixel decoding/crop/JPEG encoding, cache keys, and TTL/LRU cache implementation from root `thumbnail.go` to `internal/thumbnail`. The root file now owns only the HTTP orchestration that enforces existing public URL validation and calls the existing secured article image endpoint. No security callback, root forwarding shim, or duplication of the cache/codec policies was introduced; regression tests now use the package-owned processor and cache directly. No version or routing changes.

## Phase 24: Google News protocol boundary

Moved Google News URL identification, legacy encoded destination parsing, RPC request serialization, RPC frame parsing and signature/timestamp expressions into `internal/googlenews`. The root resolver now uses those functions directly while retaining outbound network transport, redirect limits and the existing `validatedURL` safety checks in the application. Tests call the new package directly. There are no forwarding compatibility methods or new callback-injected security dependencies. No change to v0.2.85 or existing HTTP response handling.

## Phase 25: page classification and shared source-noise ownership

Moved all five page classifiers (index, marketing landing, search portal, thin portal, sparse-root landing) out of root `page_type.go` to `internal/pageclassify`. Moved the shared noisy source-region detection (including its promotional-tag policy) out of `main.go` into `internal/sourcehtml`, so both page classification and root extraction/recovery call the same package-owned implementation directly. Removed both original implementations, updated callers and regression tests, and introduced no wrapper or callback adapter. The old parser and data models are reused, not duplicated. Application version and network access gates are unchanged.

## Phase 26: partial-excerpt extraction with shared source policies

Moved complete short-article excerpt recognition and extraction from root `partial_html_article.go` to `internal/extraction.PartialHTMLArticle` / `HasPartialArticleEvidence`, deleting the original file. To avoid duplicating policy or passing callbacks into root, moved the shared advertising-text exclusion predicate into `internal/extraction` and source URL resolution into `internal/sourcehtml`. Updated production callers and tests to reference their package owners directly, and removed the original root policy implementations. This does not fetch additional content or bypass access controls. No version, HTTP, or runtime hosting changes.

## Figure anchor regression: unmatched Readability prose

Source paragraph alignment in `positionArticleFigures` now advances only on a confirmed match with a retained text block. Missing or rewritten source paragraphs no longer exhaust the insertion cursor, which previously could push an opening editorial figure to the end of the document. Added source-order regression fixtures modeled on the Uberblok opening magnifying-glass figure and a later illustration; no publisher-specific condition or image re-ranking. This is a behavioral bug fix separate from the extraction package migration.
