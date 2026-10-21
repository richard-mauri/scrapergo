# ScraperGo v0.2.60 development patch (2026-10-21)

Use local installed v0.2.59 as baseline. Apply this direct-unzip source patch only if local diff audit shows no conflicting work. No commit or push implied.

v0.2.60 registers golang.org/x/image/webp for WebP image decoding (output remains center-cropped JPEG), adds format-specific diagnostics on HTTP 415, and includes tests for source-type detection. AVIF is not decoded and is named in error diagnostics. Existing article-image HTTP 502 causes are not addressed by this patch.

DEPENDENCY: after unzip, run `go get golang.org/x/image@v0.30.0` on the Mac (downloads module and updates go.mod/go.sum), then `go test -count=1 ./...`. Keep module files in commit after verification. This environment could not download that dependency and did not execute the full test suite. Do not advertise tested or release-ready until the local suite and live examples pass.

Leave launchers, tunnel, data, scraped articles and runtime binary alone. Changes are in main.go, thumbnail.go, thumbnail_test.go, README.md, AGENTS.md. Review git diff and do not commit automatically.

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
