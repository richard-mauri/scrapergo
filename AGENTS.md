# CURRENT WORK — PR #52 Readability source figure extraction (2026-10-23)

PR #51 merged into main at e33ed57acbb5abaf079edf55ca137fa3c31c8976. Branch refactor/extraction-figures begins there. Moves figure ordering, gallery supplementation, inline-link extraction and source image scanning from root readability.go into internal/extraction/figures.go. No root forwarding wrappers. Reader orchestration in main/readability.go remains a larger pending migration. Semantic version remains 0.2.85; build indicator Architecture PR #52. Do not merge, release, or tag without approval. Earlier notes below historical.

# CURRENT WORK — PR #51 diagnostic HTTP ownership (2026-10-23)

PR #50 merged into main at 96e98758a3ff41cdf0be8e5bf08ab8418aeb18cc. This branch refactor/diagnostics-package moves both restricted diagnostic endpoints from root into internal/diagnostics. Diagnostic handlers retain loopback enforcement, bounded JSON/HTML, no external asset requests for supplied HTML, and reuse the production extraction stage sequence through explicit dependency injection. Root tests are migrated to direct diagnostics APIs. Semantic version remains 0.2.85; display build PR #51. Extraction logic remains rooted; do not represent this PR as final architecture. Do not merge, release or tag without user approval.

# CURRENT WORK — PR #50 image delivery boundary (2026-10-23)

PR #49 merged into main at 8da00fda07f241c330d75fb7d75b3b1b9c9bd590. Branch `refactor/image-delivery-service` starts there. The secure image delivery implementation moves from root image_delivery.go into internal/imagedelivery/delivery.go; main directly composes the metadata handler into the image service and thumbnail handler. The original root module is deleted and tests migrated to package APIs without compatibility wrappers. This preserves HTTPS upgrade, validated public-host-only image URLs, redirects, 4 MiB limit, candidate fallback, safe referrer and fixed diagnostics. Semantic version stays 0.2.85, visible build Architecture PR #50. Root extraction/diagnostics and other root responsibilities still need architectural migration; do not declare bootstrap-only. No merger, release, tags, launcher or tunnel changes without owner approval. All earlier notes historical.

# CURRENT WORK — PR #49 transport and resolver package ownership (2026-10-23)

PR #48 was merged into main at 98de2afc99163049719b5de6cdb17c3de34d0766. Branch `refactor/transport-service-ownership` is based on that commit. Moves HTTP classification, thumbnail delivery and their caches into `internal/httpapi`, and Google News RPC fetch/resolution into `internal/googlenews`, deleting their old root source files. The new package-level tests cover authorization boundaries, caching, and image fallback. Explicitly wires existing read handler and validated URL policy without root compatibility wrappers. Visible build becomes Architecture PR #49; semantic appVersion remains 0.2.85. This is further extraction of cohesive root services, NOT the final bootstrap-only main package; extraction, image delivery, thumbnail and diagnostics still remain in root. No changes to launcher, tunnel, merge, tags or release without user approval. Earlier entries are historical.

# CURRENT WORK — root compatibility deletion (2026-10-23)

PR #47 merged into main at be0bd6222002967396ed49210bdbfa82578088a2. Branch `refactor/remove-root-compatibility-shims` starts from it. The task is to eliminate obsolete root compatibility-only files and migrate production/test callers to internal/recovery and internal/reader. Root implementations still include extraction, diagnostics, image serving and other delegates; the executable is NOT yet a bootstrap-only main.go. Preserve app semantics, version 0.2.85, launcher/tunnels, and do not merge/tag/release without approval. Previous AGENTS entries are historical.

# CURRENT WORK — PR #47 application routing and lifecycle (2026-10-23)

PR #46 UI separation and bridge deletion merged into authoritative main at 17232fa15f527a0a83891d979864ca031f158051. Branch `refactor/app-composition` begins from that commit. This milestone transfers complete HTTP route registration, HTML/CSP response, static portrait serving, server lifetime and startup logging into `internal/app`; production main composes legacy handlers and invokes `app.Run` rather than registering HTTP routes. Version remains v0.2.85, visible build identifies Architecture PR #47. Dedicated app tests protect route contracts, CSP, and portrait HTTP method restrictions. Existing extraction, image, and diagnostic endpoint implementations remain in root and require substantive backend migration before we can truthfully claim a bootstrap-only executable. Do not merge, tag or release without owner authorization. Earlier AGENTS handoffs below are historical.

# CURRENT WORK — PR #46 UI package and asset separation (2026-10-23)

PR #45 secure retrieval was merged into main at 1b18379bce9e2d01fbdc2fa4e1da4f1b44034b4e. Branch `refactor/browser-ui-boundary` starts from that authoritative commit. Move the browser reader's HTML, CSS and JavaScript out of the root monolithic ui.go into `internal/ui/templates/index.html` and `internal/ui/static/reader.css`, `reader.js`. Embedded assets are still rendered inline to preserve existing CSP, single-binary behavior and Safari runtime. Root `ui.go` remains a minimal compatibility bridge for existing main-package tests; it is not the presentation owner. Set visible build label to Architecture PR #46, retain semantic version v0.2.85. Do not change launcher, ports, tunnels, scraping API, article output or network access. PR #47 and subsequent work will address app composition and residual root cleanup. Do not merge, tag or release without owner approval. All prior handoffs are historical.

# CURRENT WORK — secure fetcher architecture milestone (2026-10-23)

PR #44 (visibility correction) merged into main at e44f97f6678755bc39d5beec7c39a98146296172. Branch `refactor/secure-fetcher-boundary` begins at this main commit. This is the next *architectural* milestone, now PR #45, because the visibility correction took #44. Move URL and DNS validation, redirect enforcement, transport dial restrictions, request headers, publisher denial classification and bounded HTML download into `internal/fetcher`. Root HTTP/extraction callers retain temporary wrappers for API/test compatibility; underlying logic must no longer live in `main`. Preserve security controls and retrieval behavior, port/tunnel/launcher, semantic version v0.2.85. Set visible architecture build label to PR #45. Next targets remain UI, composition, final root cleanup; the six substantial milestones target is unchanged, although PR numbering shifted due to the small visibility correction. No version release, tags, merge or widening access without owner approval. Historical sections follow.

# CURRENT WORK — PR #43 backend recovery and reader wiring migration (2026-10-23)

PR #42 merged into main at ced65130701cc334b82d65d788465874914c2214. Branch `refactor/backend-source-recovery` starts from merged main, continuing the approved six-PR architecture plan in ARCHITECTURE.md. This milestone moves source photograph recovery, source image classification, gallery grouping, opening-lead recovery, source link recovery, contributor portraits/byline classification, source image URL resolution and rendition identity into importable backend packages. Crucially, internal/reader/default.go now owns the complete production stage composition: the root document_assembler.go no longer supplies callbacks. Root functions remain small compatibility wrappers for existing extraction/test callers and are to be removed by PR #47. Preserve stage order, Hackaday featured/later images, BMW lead/gallery pair, author portraits, inline links and version 0.2.85. Next milestones: #44 network/security, #45 UI separation, #46 app composition, #47 final root removal. No new requests or releases. Owner alone merges. Earlier AGENTS entries are historical.

# CURRENT WORK — PR #42 reader architecture foundation (2026-10-23)

PR #41 merged into main at a24780c37933aad2a3742cd21691d0f854ddbea0. New branch `refactor/reader-architecture-foundation` starts from this authoritative main commit. We have approved a six-PR maximum cohesive architecture migration, documented in ARCHITECTURE.md. PR #42 establishes `internal/reader` as owner of ordered article-processing stages and assembly, with regression tests and initial prohibited-dependency checks. The root document_assembler.go currently only wires legacy callbacks and is explicitly temporary until PR #43. Future phases: #43 backend extraction/recovery, #44 retrieval/security, #45 UI, #46 app composition, #47 minimal main and final cleanup. This is a structural milestone, not an isolated helper migration. Keep appVersion 0.2.85, exact stage ordering, Hackaday opening image, BMW lead/gallery and inline links unchanged. No release, tags, new network access or merge without owner approval. Historical sections below are superseded.

# CURRENT WORK — opening lead reconciliation migration (2026-10-22)

PR #40 was verified merged into main at 23160f9a29ba1a74d05d4624b37446a0c9f338d1. New branch `refactor/opening-lead-reconciliation` begins at that main commit. Scope: move already-selected opening photograph block reconciliation (same-identity rendition promotion versus prepend) from root HTML adapter into `internal/media`. Root retains source HTML traversal, structural and noisy-container checks, URL resolution, and identity callback. Add focused tests for image ordering, rendition deduplication, and no-candidate behavior. Preserve Hackaday opening photo, BMW lead/gallery, and accepted inline links. No version bump, release, tags, new requests, startup changes, or merge without owner approval. Previous sections are historical.

# CURRENT WORK — opening-image source-noise scope cleanup (2026-10-22)

PR #39 is confirmed merged into main at 364cd9a184e2e8145fb4348c6a0589aa99401675. New branch `refactor/opening-image-noise-scope` starts at that merged commit. Focus: move the existing opening-figure noisy-ancestor rule from the root adapter to `internal/sourcehtml`, alongside the shared source-noise classifier. The narrowly scoped exception for an article's own header is preserved; byline/advertising/navigation exclusions remain. Image eligibility stays in `internal/media`; source traversal, URL selection, and insertion stay in root. Existing Hackaday, BMW lead/gallery, and link behavior must not change. No semantic app version bump, publisher handlers, extra network calls, release, or merge without owner approval. PR #39 and earlier sections below are historical.

# CURRENT WORK — opening figure policy migration (2026-10-22)

PR #38 was merged to main at e53115b5627228eaed741ee03134c284fc7cd6d3 after owner acceptance of the Hackaday opening image, BMW gallery, and link behavior. This new branch refactor/opening-figure-policy starts from that merged main and moves publisher-neutral opening-image ancestor eligibility to internal/media. Root adapter retains source DOM traversal, scope/noise checks, URL resolution, image identity matching, and document insertion. No behavior changes, publisher-specific selectors, version bump, new network operations, or launcher changes. Run Go CI and owner Mac acceptance before merging. Owner alone authorizes merges. Earlier PR #38 headers below are historical.

# PR #38 Hackaday image diagnostic — 2026-10-22

Owner accepted inline-link navigation from BMW to Hackaday but reported the opening windshield photograph absent entirely from ScraperGo; the first ScraperGo photo is a later source image. Do not misclassify as mere image positioning or assume metadata recovery. This PR adds read-only, localhost-only image trace comparison `sourceImagesAbsentFromArticle`, de-duplicated by media identity, to identify which source images never survive assembly. Next: collect Hackaday trace and use source ancestor/stage history to establish an evidence-backed generic recovery fix. The diagnostic is not a lead-selection fix; do not merge based on it alone. Preserve verified link refactor and BMW gallery. No semantic version bump.

# CURRENT WORK — PR #38 (2026-10-22)

PR #37 is confirmed merged. This branch `refactor/selected-link-assignment` begins from current main, v0.2.85 unchanged. Scope: move one-to-one matching of Readability-selected inline-link parts into `internal/links`, leaving HTML tree parsing and inline markup extraction in the root adapter. Preserve existing behavior, paragraph order, and no new network requests. New internal link matching tests. PR remains draft until CI and owner review; owner alone merges. No additional handoff files.

# CURRENT AUTHORITATIVE HANDOFF — 2026-10-22

ScraperGo v0.2.85. GitHub `main` is authoritative for each iteration; verify previous PR merged, create a new short-lived draft branch from latest `main`, run GitHub Actions tests and build, and only then offer owner Mac/Safari acceptance. The owner alone merges; do not tag, release or bump appVersion during development. This section supersedes historical instructions below mentioning `dev`, `rewrite/structured-articles` or prior draft PRs. Do not introduce root compatibility facades, callback shims or parallel handoff documents.

PR #35 also corrects first-article-image loss: internal/media.Deduplicate now deduplicates image blocks without treating Article.ImageURL metadata as an already-rendered photo. BMW Ars Technica trace demonstrated metadata rendition and editorial figure share the original image, causing its block to be deleted and a later gallery photo to become lead. Tests cover original/resized rendition plus distinct gallery. Diagnostics now share the production stage sequence. Await CI and owner Mac/Safari acceptance for both Beth Mole portrait and BMW true lead before merging.

Latest confirmed merged through PR #34 (figure anchor fix). Current PR #35, `fix/postarticle-author-portrait`, targets author-bio portraits displayed after editorial content while preserving NPR contributor portraits before editorial leads and CNN/other editorial images. PR #35 is not merged unless verified by GitHub. After acceptance, resume cleanup from updated `main`; never maintain divergent parallel cleanup and bugfix baselines. Continue the cohesive extraction/recovery migration.

# APPROVED EXCEPTION — structured extraction rewrite (2026-10-21)

The owner explicitly authorized a proper extraction architecture redesign and delegated branch management. For this project **on this branch**, make development commits to `rewrite/structured-articles` rather than the historical `dev` integration branch. This section supersedes only branch-target instructions below, not safety, CI, documentation, owner acceptance, or release gating. `main` is unchanged and remains protected from merge until owner acceptance. The complete pre-rewrite experiments remain at `archive/scrapergo-v0272-experiments`; do not cherry-pick their brittle image/embedded-markup JS into the rewrite. Keep PR #4 draft, build the frozen multi-publisher fixtures and structural image-role coverage, and validate the Mac release only after CI passes. Increment the visible version for Mac candidates and always provide full fetch, checkout, tests, build, restart, and Safari steps.

# AUTHORITATIVE DEVELOPMENT WORKFLOW — effective 2026-10-21

This section supersedes any conflicting directions in the historical handoff notes below. The notes remain for technical provenance, not as active release instructions.

## Branches, source of truth, and ownership
- GitHub `dev` is the integration branch for routine development. Before editing, read this file and `README.md`, fetch the current `dev` source, and preserve existing behavior and user state unless the task explicitly changes it.
- Commit and push focused changes directly to `dev` only while carrying out an authorized development request. Include meaningful regression tests and concise documentation; do not require the user to apply ZIP patches or hand-edit source.
- Keep `main` unchanged until the owner explicitly authorizes release promotion. Do not merge, force-push, rewrite shared history, or create a release/tag without the owner's authorization. A ready-for-review PR is not approval to merge.
- Preserve launch scripts, Cloudflare tunnels, application ports, user tokens, subscriptions/exclusions, data directories, and local configuration. Do not overwrite generated binaries or operating data.

## CI and diagnostics
- Each push to `dev` triggers `.github/workflows/go-ci.yml`; the workflow runs `go test -count=1 ./...` and `go build ./...`. Existing repository-wide formatting debt is reported without blocking these checks; format all touched Go code.
- Keep a PR from `dev` to `main` open for run discovery and job-log inspection. Commits to `dev` also trigger PR validation. Find the workflow run associated with the latest commit, inspect job steps and full logs, fix failures on `dev`, and repeat until CI passes. Do not ask the owner to monitor CI or send screenshots for ordinary compilation/test failures.
- A green build means tests compile/pass; it does not prove publisher access, Safari behavior, tunnel routing, or other live integrations. Never claim a live defect fixed solely on CI results.
- If CI or logs cannot be retrieved, explicitly report that limit and avoid claiming verification. Do not leave unverified changes represented as release-ready.

## Mac acceptance and promotion
- After CI passes, provide Mac-compatible checkout/pull, `./devbuild.sh`, and existing launcher instructions for a specific candidate. The owner's Mac uses an older Git that supports `git checkout`, not `git switch`.
- Request owner acceptance for observable changes in the live application, especially Safari/UI, RSS/publisher integrations, image retrieval, admin settings, and tunnels. Keep acceptance separate from automated CI.
- Prepare or update the PR with the intended scope, latest commit, CI outcome, live acceptance evidence, remaining limitations, and explicit no-merge-until-approved status. Only the owner decides when to approve/merge into `main`.
- After owner merges, use `git checkout main`, `git pull --ff-only origin main`, `./devbuild.sh`, and the existing application launcher to verify the promoted `main` candidate. Do not automatically merge another app's changes.
- Work occurs only during an active interaction unless a separately scheduled automation is requested; never imply continuing development after ending a reply.

# CURRENT DEV IMAGE FETCH CHANGE

Use an article-scoped Referer stripped of query/fragment and an Accept header limited to supported image formats for the selected article image request. All existing security checks remain. Tests in image_referrer_test.go. Publisher-specific image delivery remains unverified. No appVersion release bump or main merge.

# ScraperGo v0.2.65 development handoff
Classification rejects publisher-access failures as unavailable. Keep v0.2.64 image selection. Image-fetch failures are not fixed in this patch. No changes to launchers, tunnels, or runtime data. Test locally before committing.

# ScraperGo v0.2.64 development candidate — 2026-10-21

Article image metadata now selects from the SAME article image references used by the reader: `Article.ImageURL` followed by first inline `Block{Kind:"img"}` image(s), with duplicates removed. `imageUrl` remains backward-compatible and a new `imageCandidates` array describes up to the displayed article images. `/api/article-image` makes at most three secure retrieval attempts on these candidates. HTTP 502 may remain if publisher image delivery blocks every candidate; no access restrictions are bypassed. Existing SSRF, size, redirect, timeout and image format safeguards remain enforced for each request. Browser rendering and thumbnail crop/encoder are unchanged. Includes `image_selection_test.go`.

Run `go test -count=1 ./...` and `./devbuild.sh` on Mac before restart. Test Practical Engineering, Dries Buytaert, The Dial, Maevyn.codes with `/api/thumbnail`, and confirm thumbnails are 256x176 JPEG; inspect failure stage headers when not. No commits or releases performed. Local verified source always takes precedence over this staged patch.

# 0.2.63 development candidate — 2026-10-21

Classify confirmed extraction_error pages as unreadable; retain transient failures as unknown.

Install ScraperGo before NewsGo. Run full Mac tests/build and browser acceptance prior to commit. Preserve launcher/tunnel and local runtime state.

# v0.2.62 candidate — classification gate

ScraperGo POST /api/classify uses existing /api/read outcomes and caches results. Confirmed page_index becomes homepage; valid extracted articles become article; publisher errors/timeouts become unknown. NewsGo integration should follow successful validation. Classification incurs extraction cost on cache miss; it does not share a cached article extraction with thumbnails yet.

Do not commit or push without testing, and preserve local changes, launchers and tunnel configuration.

# ACTIVE HANDOFF — ScraperGo v0.2.59 corrected thumbnail candidate (2026-10-21)

- **Corrected package:** the first v0.2.59 ZIP omitted `image_delivery.go`/`image_delivery_test.go` from v0.2.58 even though `main.go` and `thumbnail.go` call `articleImage` and `newResponseCapture`. The corrected ZIP includes both dependencies. Do not build/restart using the incomplete package. Full Go tests must pass on the user Mac before deployment.
- The authoritative local baseline for this direct-unzip patch is the user-provided ScraperGo source plus the generated v0.2.58 image-delivery candidate; do not overwrite newer local edits. Earlier v0.2.56 diagnostics remain uncommitted and must be preserved.
- `GET /api/thumbnail?url=<encoded article URL>&width=256` reuses `articleImage`, which reuses `imageMetadata` and `/api/read` selection. It returns actual 16:11 JPEG thumbnail bytes, not a raw image URL. Default 256×176, width 64–640, quality 82. In-memory bounded 128-entry / 30-minute completed-thumbnail cache. JPEG/PNG/GIF input only; WebP/AVIF unsupported for now. Output cache headers are private, short-lived.
- No new third-party dependency; source-image byte cap 4 MiB remains in v0.2.58, input pixel cap 16 million. SSRF and redirects remain subject to the pre-existing image-delivery checks. Uncached requests still perform article extraction; do not claim this is safe for unthrottled public load. Test and evaluate access controls and concurrent cache misses before integrating NewsGo.
- Isolated thumbnail tests passed; full `go test ./...` blocked here by inaccessible Codeberg module. Require Mac Go 1.27.1 full suite and live CNBC / Seahorse acceptance, including content type and dimensions, prior to commit/release. Do not change NewsGo, NewsValidate, tunnel files, executable or user data in this patch.

# ACTIVE HANDOFF — ScraperGo v0.2.58 image delivery candidate (2026-10-21)

- Based on user-verified local v0.2.57 extraction metadata endpoint and still-uncommitted v0.2.56 diagnostic files. Preserve all existing local source and tests.
- Adds GET /api/article-image?url=<URL-encoded article URL>. Reuses /api/image-metadata and therefore /api/read extraction to identify the selected image, then retrieves actual image bytes with a bounded HTTPS-only image request. Caller cannot supply image URL directly. Validates destinations and redirects for SSRF, caps bytes at 4 MiB, accepts JPEG/PNG/GIF/WebP signatures, uses a bounded request deadline and no proxy. Fetches are not cached in this candidate.
- A selected image may still fail to download. CNBC browser-style request headers previously resulted in HTTP 200 on the user Mac, but publisher permissions and runtime endpoint success must be verified independently. Do not claim NewsGo integration or browser rendering completed.
- Full Go suite blocked in this environment by existing Codeberg module download. Requires go test ./..., ./devbuild.sh, ./scrapergo.sh and live endpoint acceptance using Mac Go 1.27.1 before commit. No changes to NewsGo or NewsValidate.
- This HTTP GET endpoint performs full article extraction for each uncached image request; assess cost, rate limits, abuse controls and caching before exposing broadly through NewsGo.

# ACTIVE HANDOFF — ScraperGo v0.2.57 image metadata candidate (2026-10-21)

- Based on local uncommitted v0.2.56 source, not GitHub main; preserve the existing diagnostic extraction work and do not stage the untracked root executable.
- Adds `POST /api/image-metadata` with the same JSON `{ "url": "https://..." }` input as `/api/read`; returns `url`, `imageUrl`, `imagePresent` without article blocks. Uses the existing `/api/read` handler internally, so publisher retrieval, SSRF controls, extraction, image selection and upstream errors are unchanged. This is a **small response**, not a cheaper extraction.
- No image bytes are proxied. A correct URL may still return HTTP 403 to Safari; NewsGo must verify actual browser delivery separately.
- Build/test not verified in this environment because module download is blocked. Run `go test ./...`, `./devbuild.sh`, and restart with the established `./scrapergo.sh` only after approval. Verify live CNBC output, no-image case, existing `/api/read`, and diagnostics. Do not commit/push until staged diff and tests are reviewed.
- ScraperGo and NewsGo image metadata integration is not yet implemented in NewsGo; NewsValidate remains independent.

# ScraperGo development handoff — v0.2.56

- `main.go` remains the application entry point and contains ordinary `/api/read`; appVersion is 0.2.56.
- `diagnostic_extract.go` adds `POST /api/diagnostic/extract`, requiring a direct loopback peer, JSON `{url,html}`, HTTP(S) URL metadata and <=5 MiB HTML; **no network retrieval**.
- The existing section of `api` following successful `download` moved verbatim into `writeExtractedArticle` and is used for both API handlers. Do not split article extraction policy between them.
- `diagnostic_extract_test.go` covers remote rejection, malformed payload rejection and POST-only behavior.
- Normal HTTP retry strategy is unchanged (alternative header profile remains HTTP 403 only).
- Security note: an upstream reverse proxy connecting locally can defeat the direct-peer protection. Keep this endpoint off publicly exposed proxies or tunnels.
- Direct-unzip patch root contains `main.go`, `diagnostic_extract.go`, `diagnostic_extract_test.go`, `README.md`, `AGENTS.md`. No modifications to NewsValidate or NewsGo.
- The patch was formatted with gofmt. The patch staging directory is not a full checkout, and full `go test ./...` must be run after applying to the complete repository.
