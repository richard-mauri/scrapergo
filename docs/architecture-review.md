> **Historical design record — not current instructions.** This document describes earlier migration/design states and may reference superseded branches, root package owners, aliases, and PR status. Current authority: [ARCHITECTURE.md](../ARCHITECTURE.md), [DEVELOPMENT.md](../DEVELOPMENT.md), and [HISTORY.md](../HISTORY.md).

# Architecture review and release checklist — October 2026

This review documents the state after PR #23 and scopes the final pre-release acceptance. It is not a claim that all root-package code has been migrated or that live publisher sites have been tested by CI.

## Dependency ownership

| Package | Responsibility | Direct internal dependencies |
| --- | --- | --- |
| `internal/article` | Canonical article and block data model | None |
| `internal/sourcehtml` | Legacy HTML tree, traversal, text and metadata helpers | None |
| `internal/extraction` | Extracted document and source evidence contract | `article` |
| `internal/assembly` | Ordered, copy-isolated document stages | `article`, `extraction` |
| `internal/media` | Image identity, deduplication, source role decisions | `article` |
| `internal/documentmedia` | Final editorial lead/metadata reconciliation | `article` |
| `internal/contributors` | Linked contributors and header separation | `article` |
| `internal/links` | Missing-link reconciliation | `article` |
| `internal/recovery` | Editorial continuity corroboration and replacement gate | `article`, `sourcehtml` |
| `internal/structured` | Structured inline text/formatting conversion | `article`, `sourcehtml` |
| `internal/httpapi` | Read and image-metadata HTTP adapters | `article` |

The executable `package main` still owns secure retrieval, publisher-access classification, Readability integration, DOM evidence collection, several recovery algorithms, routing and launcher wiring. It injects these functions into the HTTP and document interfaces. Do **not** describe the system as fully dependency-inverted or `main.go` as only a launcher yet.

## Preserved contracts

- `POST /api/read`: 4096-byte request cap, existing URL/Google News resolution and SSRF checks, existing access-denied and extraction/error JSON conventions
- `POST /api/image-metadata`: delegates to the same read handler and preserves lead-first candidate ordering
- Article JSON model and image role precedence, including editorial lead versus author portrait
- No additional network fetches in article assembly, contributor, media, recovery or link stages
- Existing launch scripts, port choices, Cloudflare tunnel and appVersion remain unaffected by the architecture PRs

## Test coverage and release acceptance

CI: `go test -count=1 ./...` and `go build ./...`. The synthetic source fixtures test opening lead-image selection and DOM evidence; they are **not** captures of CNN or NPR pages. New contract tests verify that a failed URL-resolution step never invokes download, oversized requests are rejected, and assembly leaves nested source blocks/links/parts unchanged.

Mac acceptance: fetch `main` after merge, run `go test -count=1 ./...`, `./devbuild.sh`, and `./scrapergo.sh`; inspect article text, links, the correct CNN editorial lead photo, NPR contributor portraits/bylines, thumbnails, and access/error paths in Safari. Only the repository owner may approve or perform merge/release promotion.

## Remaining technical debt — explicitly not closed

1. Retrieval security, network handling and article classification remain in `main.go` rather than a complete service/retrieval package.
2. Several source-image and content recovery passes still traverse shared parser trees in root code. Their ownership and testing need more work before wholesale relocation.
3. Compatibility wrappers remain in root and are used by existing tests and diagnostic paths. Remove only after caller-by-caller audit, not through bulk deletion.
4. Coverage lacks permission-appropriate archived real-world publisher HTML; live CNN/NPR acceptance is necessarily manual.
5. Broad package-migration completion does not prove the correctness of editorial ranking on every news source.

This PR is a stabilization and architecture inventory milestone, not a claim of a finished full rewrite. Further changes should be driven by a concrete dependency, code-quality or behavioral need rather than by increasing package count.
