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
