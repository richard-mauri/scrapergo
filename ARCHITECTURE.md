# ScraperGo architecture contract

The production executable is a composition root, not a processing module. ScraperGo uses one canonical transport-independent article model, and its UI, transport, orchestration, and backend services have separate ownership.

## Intended responsibility and import directions

- `main.go`: program entry point only; calls `internal/app`.
- `internal/app`: configuration, dependency composition, routing and lifecycle.
- `internal/ui`: browser assets and presentation; talks through the HTTP API, never imports extraction or recovery packages.
- `internal/httpapi`: HTTP transport and response contracts; delegates article reading to the reader service.
- `internal/reader`: article processing orchestration, ordered transformations and document assembly; no HTTP, UI or independent networking.
- `internal/fetcher`: secure network retrieval, redirect and host validation.
- `internal/extraction`, `internal/sourcehtml`, `internal/media`, `internal/links`, `internal/contributors`: cohesive backend capabilities.
- `internal/article`: canonical article and block model, independent of application and presentation layers.

Dependencies must be acyclic. Backend capability packages must not depend on `internal/ui`, `internal/httpapi`, or the root executable package. The reader must not depend on HTTP or UI. Network retrieval must remain centralized in the fetcher boundary and must not be initiated by assembly stages.

## Migration sequence (six PRs maximum)

1. #42: reader ownership of processing orchestration, architectural contract and boundary tests. Legacy callbacks remain temporarily wired by the root adapter.
2. #43: move the remaining extraction, source recovery, assembly and normalization implementations behind backend boundaries; remove the root callback adapter.
3. #44: isolate secure retrieval and publisher access policy in fetcher.
4. #45: move browser HTML/CSS/JS and rendering into UI.
5. #46: introduce app composition and move HTTP wiring, configuration and lifecycle out of root.
6. #47: remove remaining root production implementations and temporary compatibility aliases; tighten boundaries and verify the minimal main.go.

Each PR must report production root-file/line deltas, migration exceptions, package dependency changes, tests and functional contracts. Target release version remains 0.2.85 during migration. Preserve current launch commands, HTTP behavior and article/image/link behavior until an explicitly approved release. Owner approves all merges.

Passing functional tests alone does not satisfy architectural acceptance. Each milestone must move a cohesive responsibility and establish a usable package API.


## Compatibility bridge audit — PR #46

The root `ui.go` compatibility bridge is deleted during PR #46; the live HTTP handler and presentation tests use `internal/ui` directly.

The following root delegates remain technical debt, not permanent architecture:
- `document_assembler.go`: legacy `articleAssemblyStages` and `assembleArticleDocument` entry points; callers/tests should transition to `internal/reader` before the file is deleted.
- `gallery_groups.go`, `source_image_roles.go`, `source_lead_recovery.go`, `link_recovery.go`, `structured_document.go`: old source recovery, contributor and identity wrappers; migrate their remaining callers and tests to `internal/recovery`, `internal/contributors` and `internal/media`.
- `main.go`: compatibility aliases for the canonical article model, source parser helpers, secure fetcher functions and article image helpers; remove only after root extraction callers have been relocated.
- `app_wiring.go`: real HTTP wiring today, not merely a compatibility bridge. Move into application composition in the next milestone and delete it from root.

These are separate from the completed UI ownership move. Do not recreate deleted UI wrappers for test convenience. The migration finish line remains a bootstrap-only executable with no backend or presentation implementation in root.
