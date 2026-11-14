# ScraperGo architecture contract

**Status:** current merged `main` including the session-aware activity and sign-in integration. This is an ownership contract, not a claim that every implementation or test is perfect.

## Runtime and dependency direction

`main.go` is the sole production Go file at repository root. It handles version/build display, port selection and embedded Tanner photo, then calls `internal/bootstrap.Start`. It must not regain extraction, retrieval, HTTP routing, or diagnostic implementations.

`internal/bootstrap` is the production composition root: it injects the secure URL resolution and fetcher into HTTP handlers, runs article extraction/classification and assembles the handler graph. `internal/app` owns port configuration, HTTP route registration, listener and lifecycle; `internal/httpapi` owns transport contracts and adapters, including read, classification, metadata and thumbnails. `internal/ui` owns rendered HTML/CSS/JS presentation and the browser-facing reader experience.

`internal/article` owns canonical transport-independent Article, Block, TextPart and link data. `internal/reader` owns the ordered transformation pipeline and calls the actual backend implementations through `DefaultProcessor`; `internal/assembly` applies stages with document isolation. Reader code must not import `net/http`, `internal/httpapi`, `internal/ui`, `internal/app` or `internal/fetcher`. Do not introduce artificial callback shims merely to make earlier root tests compile.

`internal/extraction` owns Readability integration, extraction results/source evidence, partial-article recognition, navigation and figure extraction. `internal/sourcehtml` owns HTML parsing, traversal, metadata and source URL helpers. `internal/structured` owns serialized inline text and embedded-media cleanup. `internal/recovery` owns source-backed missing photograph/link and lead/gallery recovery; `internal/media` owns image source identity, roles, ordering and deduplication; `internal/documentmedia` owns final document/image normalization. `internal/contributors` owns bylines, contributor headers and portraits; `internal/links` owns article link policy, selection and reconciliation support. `internal/pageclassify` separates article pages from landing/index/portal pages.

`internal/fetcher` owns public URL validation, secure HTTP download, redirects and network access restrictions; `internal/googlenews` resolves Google News URLs using the existing validated retrieval path. `internal/accesspolicy` owns publisher access restrictions and readable-preview notices. `internal/imagedelivery` delivers selected article images using shared metadata/retrieval rules. `internal/thumbnail` owns thumbnail sizing, processing and cache primitives. `internal/diagnostics` owns loopback-only HTML extraction and image tracing.

Dependencies are intentional and acyclic: entry point → bootstrap → application adapters and backend services; reader → article/assembly and backend transformations; domain/backend packages never depend on UI, HTTP APIs or executable root. Only established retrieval paths may contact publishers: transformation and assembly stages must not download or bypass URL/redirect defenses. Handler code may delegate to the established validated retrieval boundary; it must not invent another fetch mechanism. Image delivery's network behavior must retain its security validation.

## Generalized article-structure preservation

Extraction must preserve reliable editorial structure from source HTML or Readability HTML through the canonical `internal/article` model and reader UI: heading levels, update/time boundaries, paragraphs, links, bylines, media and original reading order. Never flatten meaningful structural tags into plain text and then reconstruct them primarily from proximity, URLs, publisher names, a single known headline, or an English relative-time regex. Where semantics were absent in the source, use conservative, corroborated structural evidence or retain the existing safe representation without inventing a heading or date.

The generality contract applies to ordinary articles, multi-update feeds, and structurally ambiguous pages. A change must justify which source structure supports the proposed canonical blocks, why unrelated headings (e.g. sidebar/navigation) are excluded, and how fallback and duplicate-removal stages remain compatible. Publisher-specific handling is an exceptional compatibility path requiring the owner's explicit approval, a documented reason, bounded scope and generalized negative tests—not the default design.

`cmd/htmlcapture` is a standalone, opt-in local publisher HTML acquisition tool that calls the **existing validated** `internal/fetcher.Download` boundary, saving a private snapshot and metadata. `cmd/readabilityprobe` is a standalone, offline comparator of source HTML, Readability `RenderHTML` and `RenderText`; see [its README](cmd/readabilityprobe/README.md). They are diagnostic executables, not runtime services, and must not become new production network paths or weaken access restrictions. Captured publisher HTML should not be committed or uploaded without appropriate permission.

## Tests and enforcement

Unit and regression tests live beside their owning implementations. Integration tests live at the service boundary (`internal/bootstrap`, `internal/reader`, `internal/httpapi`, etc.). Two root tests remain: `architecture_boundary_test.go` and `version_test.go`. `TestProductionRootBoundary` checks sole production root `main.go`, approved root functions/imports and absent obsolete adapters. `internal/reader/architecture_test.go` enforces the reader's transport, presentation, lifecycle and retrieval import exclusions. Other package tests cover contract and behavior regressions. Current enforcement is specific, not a whole-repository import DAG validator.

No compatibility forwarding wrappers, test-only aliases, migration facades or deliberately misplaced tests. Refactoring is not complete if old adapters or tests are stranded; passing Go tests is necessary but not sufficient if ownership constraints are broken.

See [DEVELOPMENT.md](DEVELOPMENT.md) for validation and acceptance; [docs/extraction-fixtures.md](docs/extraction-fixtures.md) for synthetic fixture scope. Additional publisher HTML captures and Mac/Safari behavior still require explicit verification.

## Live-blog repeated featured updates

`internal/reader/live_updates.go` contains the conservative, post-normalization editorial text cleanup for live pages (`/live/` URL paths). It compares exact ordered textual block sequences and retains the earlier featured copy and removes only a later repeated text run, and only if a long exact match recurs. Other updates, images and timestamped later text remain. This stage performs no network access, does not rewrite ordinary article prose or remove isolated repeated quotations, and is covered by synthetic regression tests in the owning `internal/reader` package. Word count is recalculated after filtering. Publisher behavior still requires Mac Safari/live-source acceptance.

## Return to Homepage navigation
ScraperGo's browser UI exposes a visible native anchor `<a href="https://home.themauris.org/">← Back to Homepage</a>` with no `target` and no JavaScript history manipulation. It navigates in the current tab and can be used regardless of account sign-in. The link adds no new server endpoints, exposes no credentials, and does not modify app-local state. Homepage PR #12 is merged; all four launcher cards use same-tab navigation.

The suite-wide return button uses the same `.suite-home-link` presentation in all four apps: 38px minimum height, 8px/12px padding, 10px radius, muted border at rest, accent border and darker fill on hover, and a visible keyboard focus outline. Place one link at the upper-left above the primary content/heading.

The shared suite navigation style is now intentionally a quiet 12px text link directly below each page's primary heading, with subtle hover background and underline and keyboard focus outline (rather than the former 38px bordered button). NewsGo does not render a duplicate signed-out Homepage action or an app Renew control; Homepage alone manages the 12-hour login session.

## Optional Homepage activity telemetry (Issue #131, merged; verified identity added later)

`internal/activity` owns a private-key-enabled, fail-open aggregate reporter and embedded tab heartbeat, with strict same-origin `POST /api/activity/ping`, ephemeral tab ID, bounded memory, and 2/5/15/60-minute rolling counts. A 60-second server-only POST to Homepage `/api/admin/activity/ingest` uses the `scrapergo` service name, timestamp, nonce and a private bearer supplied via `SCRAPERGO_ACTIVITY_URL` and `HOMEPAGE_ACTIVITY_KEY_SCRAPERGO` environment variables. The collector URL is HTTPS or loopback-only HTTP with redirects disabled. The payload omits article URLs, user identities and identifiers, token values and publisher content. The optional telemetry client does not fetch publisher articles and cannot trigger a restart.

`internal/app` wires the reporter at process startup and wraps existing dynamic API requests to measure in-flight work; the UI homepage, photo and browser ping are not counted as extraction work. Existing `internal/httpapi`, reader, fetcher, image delivery, classification and diagnostics implementations remain unchanged. `internal/ui` preserves the current inline script CSP and appends the optional heartbeat only when the server has a valid reporter. If telemetry is unconfigured or unavailable, normal app functionality continues. Homepage-verified named/guest sessions are distinguished from anonymous tabs by server-side identity checks; invalid bearer returns 401 and verification outage 503.

For isolated Mac acceptance, `internal/app` accepts `SCRAPERGO_ACTIVITY_TEST_LISTEN=127.0.0.1:8088` (or matching `[::1]:port`) only when it exactly matches the selected port. An invalid or non-loopback override fails startup. With the override absent, the original production `:<port>` bind semantics remain unchanged. This is not a public deployment configuration and should not be set in production launchers.
