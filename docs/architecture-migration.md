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
