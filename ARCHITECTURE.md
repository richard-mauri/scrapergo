# ScraperGo architecture contract

**Status:** current merged `main` after PR #63 (2026-10-23). This is an ownership contract, not a claim that every implementation or test is perfect. See [HISTORY.md](HISTORY.md) for the completed migration.

## Runtime and dependency direction

`main.go` is the sole production Go file at repository root. It handles version/build display, port selection and embedded Tanner photo, then calls `internal/bootstrap.Start`. It must not regain extraction, retrieval, HTTP routing, or diagnostic implementations.

`internal/bootstrap` is the production composition root: it injects the secure URL resolution and fetcher into HTTP handlers, runs article extraction/classification and assembles the handler graph. `internal/app` owns port configuration, HTTP route registration, listener and lifecycle; `internal/httpapi` owns transport contracts and adapters, including read, classification, metadata and thumbnails. `internal/ui` owns rendered HTML/CSS/JS presentation and the browser-facing reader experience.

`internal/article` owns canonical transport-independent Article, Block, TextPart and link data. `internal/reader` owns the ordered transformation pipeline and calls the actual backend implementations through `DefaultProcessor`; `internal/assembly` applies stages with document isolation. Reader code must not import `net/http`, `internal/httpapi`, `internal/ui`, `internal/app` or `internal/fetcher`. Do not introduce artificial callback shims merely to make earlier root tests compile.

`internal/extraction` owns Readability integration, extraction results/source evidence, partial-article recognition, navigation and figure extraction. `internal/sourcehtml` owns HTML parsing, traversal, metadata and source URL helpers. `internal/structured` owns serialized inline text and embedded-media cleanup. `internal/recovery` owns source-backed missing photograph/link and lead/gallery recovery; `internal/media` owns image source identity, roles, ordering and deduplication; `internal/documentmedia` owns final document/image normalization. `internal/contributors` owns bylines, contributor headers and portraits; `internal/links` owns article link policy, selection and reconciliation support. `internal/pageclassify` separates article pages from landing/index/portal pages.

`internal/fetcher` owns public URL validation, secure HTTP download, redirects and network access restrictions; `internal/googlenews` resolves Google News URLs using the existing validated retrieval path. `internal/accesspolicy` owns publisher access restrictions and readable-preview notices. `internal/imagedelivery` delivers selected article images using shared metadata/retrieval rules. `internal/thumbnail` owns thumbnail sizing, processing and cache primitives. `internal/diagnostics` owns loopback-only HTML extraction and image tracing.

Dependencies are intentional and acyclic: entry point → bootstrap → application adapters and backend services; reader → article/assembly and backend transformations; domain/backend packages never depend on UI, HTTP APIs or executable root. Only established retrieval paths may contact publishers: transformation and assembly stages must not download or bypass URL/redirect defenses. Handler code may delegate to the established validated retrieval boundary; it must not invent another fetch mechanism. Image delivery's network behavior must retain its security validation.

## Tests and enforcement

Unit and regression tests live beside their owning implementations. Integration tests live at the service boundary (`internal/bootstrap`, `internal/reader`, `internal/httpapi`, etc.). Two root tests remain: `architecture_boundary_test.go` and `version_test.go`. `TestProductionRootBoundary` checks sole production root `main.go`, approved root functions/imports and absent obsolete adapters. `internal/reader/architecture_test.go` enforces the reader's transport, presentation, lifecycle and retrieval import exclusions. Other package tests cover contract and behavior regressions. Current enforcement is specific, not a whole-repository import DAG validator.

No compatibility forwarding wrappers, test-only aliases, migration facades or deliberately misplaced tests. Refactoring is not complete if old adapters or tests are stranded; passing Go tests is necessary but not sufficient if ownership constraints are broken.

See [DEVELOPMENT.md](DEVELOPMENT.md) for validation and acceptance; [docs/extraction-fixtures.md](docs/extraction-fixtures.md) for synthetic fixture scope. Additional publisher HTML captures and Mac/Safari behavior still require explicit verification.
