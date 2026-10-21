## Structured article rewrite — v0.2.73 development candidate

This rewrite branch is based on current main; it is not a promoted release. Instead of importing the v0.2.66–v0.2.72 display heuristics, it preprocesses the HTML DOM to remove interactive media nodes before Readability, strips serialized media tags from ordinary text nodes while retaining code examples, and uses the same sanitized HTML for content, image and link recovery. After extraction, document normalization reconciles a duplicated opening title and constructs one typed image-block stream (optional role `lead`) for the UI. The existing `imageUrl` response remains available as metadata for API compatibility. The approved NewsGo-like dark palette and centered 1200px layout are retained.

Automated checks cover media removal, surrounding prose, retained code samples, lead vs. author images, single rendering path, the visual palette, and an end-to-end synthetic article response. These checks do not yet establish that all live NPR article image variants are deduplicated or that publisher HTML renders correctly in Safari and print preview. The rewrite remains draft until representative frozen publisher fixtures and Mac acceptance pass. Do not merge into main without explicit owner approval.

## Development candidate: image HTTP request compatibility (dev)

The image delivery fetch now sends the full article URL (without query or fragment) as Referer rather than the publisher site root, and advertises only JPEG/PNG/GIF/WebP formats that the thumbnail decoder supports. This preserves the existing validated image selection, HTTPS-only/public-host requirements, redirect checks, response size and deadline limits. Regression tests cover article-scoped, query-stripped referrers. These request changes may help legitimate publisher hotlink checks, but real Practical Engineering and The Dial 502 failures remain unverified; no live success is claimed. Do not merge to main before Mac acceptance tests.

## v0.2.65 — conservative classification (development)

`POST /api/classify` now explicitly marks publisher-access and transport failures as `unavailable` instead of `unknown`. NewsGo v0.2.52 displays only confirmed `article` results. This avoids headlines that open to HTTP 202 (Accepted) publisher responses. Unavailable results expire after one minute and do not permanently exclude subscriptions. Existing article extraction, thumbnail selection, secure image fetching, and JPEG output are unchanged. Image-fetch failures remain unresolved; no new shared extraction cache is claimed. Run full Mac tests and live acceptance before committing.

# ScraperGo v0.2.64 development candidate — 2026-10-21

Article image metadata now selects from the SAME article image references used by the reader: `Article.ImageURL` followed by first inline `Block{Kind:"img"}` image(s), with duplicates removed. `imageUrl` remains backward-compatible and a new `imageCandidates` array describes up to the displayed article images. `/api/article-image` makes at most three secure retrieval attempts on these candidates. HTTP 502 may remain if publisher image delivery blocks every candidate; no access restrictions are bypassed. Existing SSRF, size, redirect, timeout and image format safeguards remain enforced for each request. Browser rendering and thumbnail crop/encoder are unchanged. Includes `image_selection_test.go`.

Run `go test -count=1 ./...` and `./devbuild.sh` on Mac before restart. Test Practical Engineering, Dries Buytaert, The Dial, Maevyn.codes with `/api/thumbnail`, and confirm thumbnails are 256x176 JPEG; inspect failure stage headers when not. No commits or releases performed. Local verified source always takes precedence over this staged patch.

# 0.2.63 development candidate — 2026-10-21

Classify confirmed extraction_error pages as unreadable; retain transient failures as unknown.

Install ScraperGo before NewsGo. Run full Mac tests/build and browser acceptance prior to commit. Preserve launcher/tunnel and local runtime state.


## Development candidate v0.2.62 — 2026-10-21

ScraperGo POST /api/classify uses existing /api/read outcomes and caches results. Confirmed page_index becomes homepage; valid extracted articles become article; publisher errors/timeouts become unknown. NewsGo integration should follow successful validation. Classification incurs extraction cost on cache miss; it does not share a cached article extraction with thumbnails yet.

## v0.2.59 corrected thumbnail ZIP (2026-10-21)

The original v0.2.59 distribution omitted `image_delivery.go` and `image_delivery_test.go` needed by the thumbnail endpoint. The corrected package includes those v0.2.58 files alongside the v0.2.59 changes. Run `go test ./...` before building or restarting. No publisher data or runtime configuration is bundled.

## v0.2.59 development candidate — ScraperGo thumbnail API (October 21, 2026)

`GET /api/thumbnail?url=<URL-encoded article URL>&width=256` returns a **real resized JPEG thumbnail** using the existing `/api/article-image` pipeline, which itself reuses article extraction and publisher image delivery. The default width is 256 pixels; allowed widths are 64–640 pixels. Output is center-cropped 16:11 (256×176 by default), JPEG quality 82, for NewsGo's existing 128×88 CSS card footprint at 2× density. No new image parsing or publisher-facing logic is added to NewsGo.

A bounded process-local cache stores up to 128 completed JPEG thumbnails for 30 minutes, keyed by article URL and width; repeated requests can be served without re-extraction or image fetching. No persistent image files or credentials are written. On a cache miss, full extraction and image download remain expensive. This initial dependency-free implementation decodes JPEG, PNG and GIF; WebP/AVIF and other source formats return an unsupported-image error rather than masquerading as resized images. The 4 MiB downloaded-image cap and public HTTPS/SSRF checks from v0.2.58 still apply; input pixel count is capped at 16 million. A successful output image is not proof that all publishers permit automated delivery. The endpoint has not been live-tested against CNBC, Safari, or NewsGo. Add controlled endpoint-rate limits / cache coalescing before exposing broadly to untrusted public traffic.

**Verification:** formatting and isolated thumbnail implementation tests passed in this environment; the complete ScraperGo suite could not run because the existing Codeberg Readability dependency was unavailable. On the user Mac with Go 1.27.1, run `go test ./...`, `./devbuild.sh`, `./scrapergo.sh`, then verify actual CNBC and Seahorse thumbnails, dimensions, cache HIT/MISS headers, and unchanged `/api/read` and diagnostic extraction. Do not commit/push before these checks.

## v0.2.58 development candidate — Article image delivery

`GET /api/article-image?url=<URL-encoded article URL>` returns the actual image bytes selected by existing article extraction (rather than merely an `imageUrl`). It calls the v0.2.57 metadata handler, which reuses `/api/read`, then fetches only that selected image. The response has an appropriate image MIME type or an error status. HTTPS image URLs, public hosts and redirects, a 4 MiB byte limit, timeouts, no upstream proxy and decoded image signatures constrain retrieval. This is not an arbitrary image URL proxy. Requests currently redo article extraction and image retrieval and are not cached; verify performance and security before broad exposure. The image URL may still be denied by publisher servers; no success is implied by metadata discovery. NewsGo and NewsValidate remain unchanged. The Go test/build suite must run locally on Mac Go 1.27.1 before restart or commit.

## v0.2.57 development candidate — Image metadata API (not yet verified)

`POST /api/image-metadata` accepts the same JSON article URL input as `/api/read` and returns only `{ "url": "...", "imageUrl": "...", "imagePresent": true }` (or false with an empty image URL). It delegates to the normal `/api/read` extraction handler and forwards structured error responses and status codes. It avoids duplicating extraction logic, but it **still runs the complete article retrieval and extraction**, so the response is smaller but the processing cost is not reduced. This endpoint does not download, proxy, or guarantee browser access to image bytes. The CNBC CDN may deny image requests depending on headers. Requires full local Go tests and live acceptance before commit or deployment. Preserves v0.2.56 diagnostic endpoint and uncommitted files.

# ScraperGo — Clean Web Reader

**Current GitHub source baseline:** v0.2.49 (October 19, 2026), commit `c24347c574ff638f29acc261f5ee46a459ea2363`. **Current development patch:** v0.2.55 unreleased anchored-tail extraction candidate (October 20, 2026); not pushed, tagged, deployed, or validated on the user's Mac. Do not assume the GitHub tree matches local uncommitted changes.

**Repository:** https://github.com/richard-mauri/scrapergo  
**Public reader:** https://scrapergo.themauris.org/  
**Local project:** `~/projects/scrapergo`  
**HTTP port:** 8083

## Purpose and architecture

ScraperGo is a self-hosted Go clean reader for publicly accessible HTTP(S) articles, used directly in Safari and as NewsGo's article reader. It retrieves server-accessible HTML, extracts text and images, and renders a clean reading view with article metadata, inline links, and access diagnostics. It cannot bypass subscriptions, logins, publisher blocking, or JavaScript-only article content. NewsGo is an independent RSS aggregator at https://github.com/richard-mauri/newsgo; a ScraperGo extraction fix must not change NewsGo.

Files: `main.go` manages version, HTTP retrieval, URL security, HTML parsing and fallback extraction; `readability.go` integrates Go Readability v2; `structure.go` recovers multi-heading editorial sections; `leading_recovery.go` is an unreleased candidate for missing initial paragraphs; `ui.go` renders safe browser output; `safari-extension/` contains the independent Safari extension; `scrapergo.sh` starts the local application and its tunnel; `devbuild.sh` builds and optionally releases executables. Root tests exercise parsing, reader behavior, images, inline links, publisher access, and navigation.

## October 20, 2026 — unresolved missing introductory paragraphs

The user reported a concrete failure in the Ars Technica article:

https://arstechnica.com/cars/2026/10/visiting-slates-first-showroom-as-its-25000-electric-truck-nears-production/

The original article starts with “As I sat in the cab of the pickup truck ...”. The ScraperGo reader begins only at “It was bigger than I expected—to my eye, roughly the size of a Hyundai Santa Cruz.” Eight preceding paragraphs were reported missing; two inline links occur within the missing content. This is a user-observed failure boundary, not proof that a particular source selector is responsible. The issue belongs to ScraperGo extraction, not NewsGo feeds.

Current source `readability.go` accepts 60+ words of Readability output without determining whether the leading source paragraphs have been dropped. The existing `structure.go` multi-heading recovery requires at least five editorial headings and may not help an article with fewer headings.

**Patch candidate:** `leading_recovery.go` uses the already downloaded source HTML and the first substantial retained paragraph as an anchor. It searches within the enclosing `<article>` (or, conservatively, `<main>`), prepends eligible preceding paragraphs in source order, rejects noisy/promotional or duplicate content, and retains inline hyperlinks. It declines recovery if the anchor or editorial scope cannot be found. The implementation is publisher-independent. `leading_recovery_test.go` covers eight missing paragraphs, order, an inline link, and a no-scope case. The `readability.go` hook runs after structured extraction and before the 60-word threshold, recalculating the word count. No claims are made that the live Ars page has passed this regression: the supplied fixture is synthetic, and the user's live reader has not been retested.

**Packaging correction (October 20, 2026):** The previously generated leading-paragraph patch included an `apply-patch.sh` helper and a nested ZIP directory. That archive does **not** meet the established minimal installation workflow and should not be installed. The leading-paragraph recovery remains an **unverified development candidate**, not an installed change. The v0.2.50 direct-unzip source candidate now contains the complete updated `main.go`, `readability.go`, `leading_recovery.go`, `leading_recovery_test.go`, and this updated root `README.md` at archive root, with no installation script and no Python dependency. Before producing that replacement, reconcile any local changes to `readability.go` to avoid overwriting a newer working copy.
**Acceptance checks:** The real Ars article should begin with “As I sat in the cab ...” and include all nine user-supplied paragraphs in order, ending that introductory span with “It was bigger than I expected ...”. Verify both embedded hyperlinks, image placement/deduplication, the rest of the story, other publishers, and absence of duplicated or non-editorial paragraphs. If the upstream retrieved HTML lacks those paragraphs, recovery cannot recreate them and the failure must instead be investigated at retrieval. Test actual browser output and access diagnostics before any release.

## Deployment

The server Mac runs an Intel executable on `http://localhost:8083/`; the public HTTPS reader at https://scrapergo.themauris.org/ uses an existing Cloudflare Tunnel. The existing `scrapergo.sh` manages the installed application and tunnel and must not be overwritten by narrow source patches. A separate Safari Mac or extension can use the HTTPS reader URL; local Safari may use localhost. Updating a Go binary does not reload Safari extension code. A tunnel is not itself an authentication layer; preserve tunnel secrets and Cloudflare access controls.

## Development Workflow and ChatGPT Handoff

ScraperGo uses a **minimal direct-unzip source patch workflow**. Every subsequent source-code patch must include the complete updated root `README.md`, along with only the changed Go source and test files, at the **ZIP archive root** (no enclosing directory). Never require Python, a separate patch application shell script, a separate ZIP inspection command, or an initial `git status` for normal patch installation. The canonical local project directory is `~/projects/scrapergo`; the standard downloaded patch resides in `~/Downloads/`. Use the actual issued ZIP filename in place of `<patch-name>.zip`:

```sh
cd ~/projects/scrapergo
unzip -o ~/Downloads/<patch-name>.zip
go test ./...
./devbuild.sh
./scrapergo.sh
```

`./devbuild.sh` runs the project build workflow and produces Intel (`darwin/amd64`) and Apple Silicon (`darwin/arm64`) standalone executables under `dist/`. `./scrapergo.sh` is the existing, verified application-and-Cloudflare-Tunnel startup script on the server Mac; it must not be overwritten by patches. Restart through the existing operational procedure, ensuring any prior ScraperGo process is stopped as appropriate; the startup script must not be assumed to terminate an already-running instance. Confirm local `http://localhost:8083/` and the public `https://scrapergo.themauris.org/`, then test the affected article in Safari and the NewsGo-to-ScraperGo reading path. For the current defect, verify the Ars Technica article begins “As I sat in the cab ...”, retains all eight formerly missing paragraphs before “It was bigger than I expected ...”, preserves inline links and image placement, and still reads cleanly through the end. Verify other publishers and existing regression behavior too.

The ZIP workflow intentionally has **no routine preliminary Git commands**. Local verified source takes precedence over GitHub `main`; when the current local file differs or contains uncommitted changes, reconcile that baseline **before** generating a replacement ZIP, not by complicating its installation procedure. Do not clear browser website data or overwrite existing tunnel/startup scripts.

### Separate review and release procedure

**Only after successful tests, build, restart, and browser verification** should Git review, commit, push, and tagging or publication occur. The application version is advanced in both `main.go` and this README for each identifiable installable build. Review `git status`, `git diff --check`, and `git diff`; stage and commit only intended changes and push the verified state. If preparing a release, verify the published version/tag is unused and the tree is clean; `./devbuild.sh --release vX.Y.Z` requires a pushed commit and an authenticated `gh`. Do not silently commit, push, tag, or create a release during patch regeneration. The current installable development candidate identifies itself as v0.2.50; no v0.2.50 release tag or publication is implied.

## Prior verified development milestones

- v0.2.9 — Go Readability integration and retrieval improvements.
- v0.2.14 — images, direct Safari reading, access diagnostics.
- v0.2.17 — Safari tab reuse and original README baseline.
- v0.2.21 — hyperlinks, gallery images and captions.
- v0.2.33 — generic HTML recovery and cleanup.
- v0.2.35 — homepage recognition.
- v0.2.40 — improved publisher access handling.
- v0.2.41 — malformed-HTML parser panic fix.
- v0.2.42 — Tanner in the header.
- v0.2.46 — inline image support and deduplication.
- v0.2.49 — hyperlinks preserved at their reading positions; baseline release.

## Permanent handoff and regeneration conventions

This root README is the single cumulative project development and chat handoff document. Update it with every regenerated source patch; do not generate independent `SCRAPERGO-v*.md` notes. Supply a focused ZIP patch containing changed source/test files and the full root README. Local verified code is authoritative over older GitHub sources. At a new-chat handoff establish exact baseline, verified tests, live behavior, outstanding problems and the next action; when local changes exist, request affected current files/source archive before replacing whole files.

Protect established behavior across all publishers: safe hyperlink rendering, image placement and deduplication, access notices, tab reuse, malformed HTML handling, and independent tunnel operation. Do not overwrite startup scripts, secrets, Cloudflare configuration or unrelated code. Distinguish tests actually executed from user-reported tests and from mere source review.

**Current handoff:** The earlier scripted, nested-directory leading-paragraph ZIP was rejected for not following the installation convention. No GitHub change, local installation, full suite, live Ars comparison, or macOS build has been performed here. v0.2.49 remains the latest verified pushed release baseline. This regeneration supplies a complete, direct-unzip, archive-root ZIP with the changed Go files, tests, and this README. It remains unverified against live publisher HTML. Obtain/reconcile newer local files first if they differ from GitHub. Then follow the minimal installation, testing, build/restart and behavioral validation workflow above. The present README-only handoff update is documentation, not proof that the extraction defect is fixed.

## Leading-paragraph recovery regeneration (October 20, 2026)

A direct-unzip source candidate now adds `leading_recovery.go` and `leading_recovery_test.go` and updates the full `readability.go`. The generic recovery anchors to the first substantive paragraph retained by Readability inside an article/main subtree, then prepends preceding editorial paragraphs, retaining inline anchors in their original positions. It declines uncertain scopes and avoids duplicating paragraphs. The minimum threshold for candidate opening paragraphs is 25 characters so short attributed quotations are not silently dropped. This candidate sets `appVersion = "0.2.50"` in `main.go`, visible in the browser UI, and records v0.2.50 here. This candidate is not a verified fix: actual recovered content depends on publisher HTML containing the omitted text; the full project tests, live Ars comparison, and Mac build/restart remain to be completed. This ZIP is based on the pushed GitHub v0.2.49 source; reconcile any newer local `readability.go` changes before applying.

## v0.2.50 development candidate — October 20, 2026

This installation candidate is based on the verified v0.2.49 `main.go` source supplied by the user, identical to GitHub `main.go`. `main.go` now declares `appVersion = "0.2.50"`, which the existing `ui.go` renders in the browser. It includes the generic leading-paragraph recovery hook and regression tests, with no publisher-specific text inserted in production code. The candidate is not a tagged release; Go test results on the full project, macOS build, reader restart and real Ars Technica HTML validation still must be performed on the user’s Mac. The canonical five-command direct-unzip workflow above remains the installation procedure.

## v0.2.52 — Mixed-layout article continuity candidate (October 20, 2026)

This development candidate extends the v0.2.51 continuity recovery beyond a strict
`<article>` scope to appropriate `<main>` and article-body containers. Traversal
distinguishes ad-only regions from wrappers that also contain substantial editorial
paragraphs; it retains neighboring prose and embedded figures. Figure captions
are emitted as image captions rather than duplicated ordinary paragraph blocks.
A new synthetic regression test covers paragraphs surrounding advertising, an
inline gallery, a final editorial paragraph, and no duplicated image caption.

This is **not yet a verified fix for the live Ars Technica Slate showroom URL**:
it depends on the publisher returning the missing editorial HTML to ScraperGo,
and full project tests plus the live reader/TXT comparison still need to be
performed on the established Mac checkout. In particular, if the publisher
response contains only the prefix through the Squareback/Fastback customization
paragraph, no HTML-only recovery algorithm can invent the absent continuation.

### Development Workflow and ChatGPT Handoff

For an already verified project at `~/projects/scrapergo`, save a direct-unzip
patch ZIP to `~/Downloads` and use the existing verified commands:

```sh
cd ~/projects/scrapergo
unzip -o ~/Downloads/scrapergo-v0.2.52-article-continuity.zip
go test ./...
./devbuild.sh
./scrapergo.sh
```

Verify the version displayed in Safari is v0.2.52. Test both the article
reading view and its TXT download against the original. The opening starts
with "As I sat in the cab"; the final paragraph should end "Now it’s time
to deliver." Verify the two-image gallery and its captions, image order,
inline links, and absence of duplicate captions or advertising.

The ZIP contains changed complete Go source/test files and this complete
cumulative README at archive root. No Python, installer, ZIP inspection,
or routine preliminary `git status` is required to install a patch.
Git review, commit, push, tagging and publication are a **separate release
procedure** only after successful testing.

## v0.2.53 — Mixed advertising wrappers and gallery recovery (development candidate; October 20, 2026)

The v0.2.52 candidate failed `TestEditorialContinuityMixedWrapperAndCaptionOnce`, returning the two originally selected paragraphs instead of recovering the gallery and closing paragraphs. The failure was reproduced in an isolated Go harness. Two code-level causes were identified: `sentenceAdRE` discarded a legitimate long paragraph merely because it mentioned “advertisement,” and the image path rejected figures under any `noisy` ancestor, including a mixed advertising/editorial wrapper that the parent traversal correctly allowed. As a result, the candidate lacked the additional text/images needed to qualify for replacement.

v0.2.53 changes `continuity_recovery.go` to suppress promotional phrases in this recovery pass only for short copy (under 12 words), while preserving substantive prose that incidentally mentions advertising. It also removes the redundant image-level `inNoisy` check; ad-only subtrees remain pruned by the existing traversal guards, and editorial figures in mixed wrappers are retained. `continuity_recovery_test.go` additionally asserts that the third editorial paragraph survives. `main.go` sets `appVersion = "0.2.53"`, visible in the UI. This patch assumes v0.2.52 is already installed, preserving its earlier `readability.go` hook and `leading_recovery.go` files.

Verification: three regression tests from the supplied continuity test file passed in a **standalone diagnostic harness with locally reconstructed dependencies**. This is not equivalent to `go test ./...` in the complete ScraperGo repository, which has not been run here. Live Ars Technica HTML retrieval and end-to-end article comparison remain unverified. Do not tag or release until full local tests and visual/TXT checks pass.

### Development Workflow and ChatGPT Handoff — installation

The normal development patch installation remains five commands, with no Python, extra patch-install scripts, ZIP inspection requirement, or preliminary `git status`:

```sh
cd ~/projects/scrapergo
unzip -o ~/Downloads/scrapergo-v0.2.53-mixed-wrapper-fix.zip
go test ./...
./devbuild.sh
./scrapergo.sh
```

Only after test/build/restart and browser article checks succeed should Git review, commit, push and release tagging be considered as a separate release procedure. Update this cumulative root README in every subsequent source-code patch.

## v0.2.54 — Retrieval versus extraction diagnostic (October 20, 2026)

The live Ars Technica Slate showroom reader still ends after the Slate Garage
customization paragraph in v0.2.53, omitting final paragraphs and gallery.
A direct `curl -L -s` request returned HTTP 403, which does not establish
what ScraperGo's HTTP client receives. Synthetic continuity tests did not
reproduce the live failure. Do not assume extraction is at fault without
inspecting ScraperGo's downloaded HTML.

**Diagnostic only:** `main.go` now calls `logSlateRetrievalDiagnostic`
immediately after a successful ScraperGo download, before any extraction,
and `logSlateExtractionDiagnostic` after Readability and recovery processing.
New `slate_diagnostic.go` makes no network calls and changes no extraction
logic or article responses. It limits logging to the reported Slate showroom
URL. The log lines `slate_diagnostic stage=download` and
`slate_diagnostic stage=extraction` report byte/block/image counts and boolean
presence of the Slate Garage paragraph, RepairPal paragraph, final sentence,
and the two missing gallery filenames. No raw HTML, article text, or
credentials are logged. Raw HTML matches may originate in scripts/JSON;
compare `raw_*` and `visible_*` markers before diagnosing.

Interpretation: HTML contains ending but extracted blocks do not => extraction
or recovery; HTML lacks ending => retrieval or separately rendered source.
If the publisher denies the request, the existing access response applies and
no successful-download diagnostic line will appear.

Patch root contains complete `main.go`, `slate_diagnostic.go`,
`slate_diagnostic_test.go`, and complete cumulative `README.md`.
Assumes v0.2.53 is already installed. Existing recovery sources unchanged.

```sh
cd ~/projects/scrapergo
unzip -o ~/Downloads/scrapergo-v0.2.54-diagnostic.zip
go test ./...
./devbuild.sh
./scrapergo.sh
```

Reopen the article and copy the two `slate_diagnostic` log lines from the
ScraperGo terminal. Confirm the browser version is v0.2.54. No commit,
push, tag, or release is part of this diagnostic patch. Remove logging after
completing the investigation.

## v0.2.55 — Anchored editorial continuation (October 20, 2026)

User's v0.2.54 diagnostics conclusively show ScraperGo receives HTTP 200
and full visible HTML: `html_bytes=182015`, `visible_repairpal=true`,
`visible_closing=true`, `raw_pegboard=true`, and
`raw_interior_models=true`. Yet final extraction reports 24 blocks, 3 images
and all four corresponding markers absent. Network retrieval is not the cause.

`continuity_recovery.go` now falls back to an **anchored tail recovery** when
full-subtree replacement cannot be independently verified. It locates the last
substantial Readability paragraph uniquely in original HTML, requires the
same article/main scope, then walks subsequent source nodes in document order,
retaining prose and captioned images while pruning isolated ads and publisher
UI. It keeps already-selected introductory blocks in place, avoids exact
text/image duplicates, preserves inline links, and refuses the extension if
fewer than two substantive additional paragraphs qualify. A synthetic
regression test covers a mixed ad/editorial wrapper, gallery and final prose.
This algorithm has NOT yet been validated against actual Ars HTML; other
markup patterns may require refinement. Diagnostics from v0.2.54 remain active
and will show whether the recovered final content reaches ScraperGo blocks.

The direct-unzip ZIP contains `main.go`, `continuity_recovery.go`,
`continuity_recovery_test.go`, and this complete `README.md` at ZIP root.
Assumes v0.2.54 is installed and retains its `slate_diagnostic.go` sources.

```sh
cd ~/projects/scrapergo
unzip -o ~/Downloads/scrapergo-v0.2.55-anchored-tail.zip
go test ./...
./devbuild.sh
./scrapergo.sh
```

Confirm v0.2.55 and check the downloaded TXT through the final sentence,
and verify two later gallery images and caption deduplication. Do not push,
commit, tag or release before live tests succeed. This remains an unreleased
candidate; an isolated fixture alone does not certify the live article.

## v0.2.56 — localhost diagnostic HTML extraction

Normal `POST /api/read` fetching, HTTP fallback, extraction and browser UI remain unchanged in behavior. The separate `POST /api/diagnostic/extract` receives JSON with `url` and `html` and passes the supplied HTML through the **same post-download extraction routine** as `/api/read`, including article/index checks, restricted-preview classification, continuity and link/image recovery. It never fetches the supplied URL or HTML-linked assets.

The endpoint requires a direct loopback TCP peer (`127.0.0.1` or `::1`), accepts only HTTP(S) URL metadata, limits HTML to 5 MiB, and returns the normal article JSON or the existing extraction error JSON. This loopback gate is not authentication; do not expose it through a reverse proxy or tunnel that forwards connections over loopback. Requests are not persisted by the endpoint. Avoid submitting untrusted HTML to a public server.

Example from the ScraperGo host, using JSON constructed from NewsValidate's saved HTML:

```sh
python3 -c 'import json,sys;print(json.dumps({"url":sys.argv[1],"html":open(sys.argv[2]).read()}))' \
  'https://example.com/article' 'publisher-response.html' | \
  curl -sS http://127.0.0.1:8083/api/diagnostic/extract \
    -H 'Content-Type: application/json' --data-binary @-
```

The sample above illustrates the request format; do not install Python solely for this operation. NewsValidate can submit JSON directly using its existing Go HTTP client. The normal `/api/read` route remains the default; the diagnostic route is opt-in.
