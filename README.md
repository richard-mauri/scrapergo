# ScraperGo — Clean Web Reader

**Current source version:** v0.2.49 (confirmed in `main.go` on GitHub `main`, October 19, 2026).
**Latest verified GitHub commit:** `c24347c574ff638f29acc261f5ee46a459ea2363` — Release ScraperGo v0.2.49, preserve article hyperlinks.
**Repository:** https://github.com/richard-mauri/scrapergo
**Public reader:** https://scrapergo.themauris.org/
**Server port:** 8083
**Local project:** `~/projects/scrapergo`

## Purpose and capabilities

ScraperGo is a self-hosted Go clean-reading service for HTTP(S) articles from many websites, not just news. It retrieves accessible HTML, identifies the article content, and presents a simplified reading view with text, images, and hyperlinks. Google News link resolution is an additional capability, not a prerequisite.

ScraperGo is intended for reading, not for archiving or bookmarking. It cannot bypass paywalls, authentication requirements, publisher access restrictions, or content only rendered client-side with JavaScript. Non-article pages, interactive sites, and catalogs may be unsuitable for extraction.

ScraperGo also serves as the article reader for the companion RSS aggregation app, [NewsGo](https://github.com/richard-mauri/newsgo). The two repositories, local services, and Cloudflare Tunnel routes are independent; do not modify NewsGo when making a ScraperGo-only change unless an integration change is explicitly required.

## Deployment configurations

**Server Mac with Cloudflare Tunnel:** ScraperGo listens locally at `http://localhost:8083/`. An existing `cloudflared` tunnel exposes the reader at `https://scrapergo.themauris.org/`. HTTPS terminates at Cloudflare; the local ScraperGo endpoint uses HTTP.

**Remote Safari client:** Safari and the temporary Safari extension run on another Mac, with the extension's Reader URL set to `https://scrapergo.themauris.org/`. The remote Mac does not need Go, ScraperGo binaries, or a local tunnel.

**Local-only Safari:** Safari, extension, and ScraperGo share one Mac; the extension's Reader URL is `http://localhost:8083/`.

A Cloudflare Tunnel is not, by itself, an authentication system. Public access to the reader's URL-fetch endpoint creates an exposure; use appropriate Cloudflare Access policies and rate limits where necessary. Never include tunnel tokens, private configuration, or credentials in patches or releases.

## Running and building

From `~/projects/scrapergo`:

```bash
go test ./...
./devbuild.sh
```

`devbuild.sh` runs the tests, builds standalone Intel (`darwin/amd64`) and Apple Silicon (`darwin/arm64`) executables into `dist/`, and produces `dist/SHA256SUMS`. The Go module currently declares Go 1.23.2. Use an appropriate installed Go toolchain.

To start the built Intel reader directly:

```bash
./dist/scrapergo-darwin-amd64 -port 8083
```

On an Apple Silicon machine substitute `scrapergo-darwin-arm64`. Check the local service at `http://localhost:8083/` before troubleshooting the tunnel.

On the established Intel host, `./scrapergo.sh` is the existing application-and-tunnel startup script. It uses local paths and secrets to launch only ScraperGo's executable and tunnel; preserve it and do not overwrite it in unrelated source patches. The startup script expects the Intel executable in `dist/` and existing local tunnel credentials.

After tests pass and changes have been committed and pushed to GitHub `main`, the existing script can publish release binaries with:

```bash
./devbuild.sh --release vX.Y.Z
```

Publishing requires an authenticated GitHub CLI (`gh`), a clean tracked working tree, a pushed branch tip, and a version/tag that does not already exist. Do not reuse published release tags. Source-only README changes should not automatically trigger a new executable release.

## Safari extension

The source tree includes `safari-extension/`; follow its [installation instructions](safari-extension/README.md). Right-click a website link and choose **Open Link in ScraperGo**. The extension navigates to the reader without opening an intermediate website tab, and reuses an existing ScraperGo tab when possible.

For a remote client, set Reader URL to `https://scrapergo.themauris.org/`; for local-only operation, use `http://localhost:8083/`. Reload the temporary extension after changing its JavaScript or manifest. Go server rebuilds do not automatically reload the extension.

The GitHub release binaries and checksums do not include the extension source; obtain `safari-extension/` from the repository.

## Technical architecture

- `main.go`: application version, HTTP handling and retrieval, Google News link resolution, HTML parsing and extraction fallback, text/image/link block models, access diagnostics, and input/host safety checks.
- `readability.go`: integration with Go Readability to recover article content.
- `ui.go`: browser reading interface and article block rendering.
- `safari-extension/`: Safari browser integration; separate reload lifecycle.
- `devbuild.sh`: tests, architecture-specific macOS builds, checksums, and optional GitHub release publishing.
- `scrapergo.sh`: deployment-specific local startup and tunnel handling; do not casually replace.

The article representation includes metadata (title, byline, publication information, source URL, word count and access notes) and ordered content blocks. Current text parts retain inline anchor targets, emphasis, and reading position. Images are represented as content blocks; source-relative URLs are resolved against the article URL.

HTML extraction uses general-purpose heuristics for removing advertising/navigation clutter and identifying article containers, with alternative content recovery where appropriate. Implement fixes at the generic extraction layer where possible; avoid publisher-specific exceptions that disrupt other sites.

## Development history and verified milestones

This section is cumulative. Dates and titles below are grounded in GitHub commits; detailed behavior must be confirmed from source and regression tests before changing it.

| Version | Development milestone |
| --- | --- |
| v0.2.9 | Go Readability integration and improved HTTP retrieval. |
| v0.2.14 | Image support, direct Safari reading, and access diagnostics. |
| v0.2.17 | Safari reader tab reuse; original setup README baseline. |
| v0.2.21 | Preserve hyperlinks, gallery images, and captions. |
| v0.2.33 | Generic HTML recovery and cleanup. |
| v0.2.35 | General-purpose homepage detection. |
| v0.2.40 | Improved publisher access handling. |
| v0.2.41 | Malformed HTML parser panic correction. |
| v0.2.42 | Tanner added to header. |
| v0.2.46 | Inline images and deduplication. |
| v0.2.49 | Preserve article hyperlinks in the extracted reading view. |

Important current behaviors: inline hyperlinks remain clickable and occur in their original reading positions; inline photographs are preserved where possible, with image deduplication; malformed HTML must not crash the application; inaccessible sources should produce meaningful access diagnostics rather than pretend a full article was retrieved.

## Testing and regression precautions

Before changing or packaging source, inspect the current local files: the GitHub `main` branch may be behind uncommitted local development. Run `go test ./...` and `./devbuild.sh` where the required Go toolchain and dependencies are available. Also check the browser UI and Safari extension separately for changes affecting display or navigation. Clearly distinguish automated tests that actually ran from source inspection or manual tests not performed.

Protect the following in every update: inline links and their targets, image ordering/deduplication, access diagnostics, normal extraction from many publishers, Safari tab reuse, independent tunnel operation, and handling of malformed HTML. Don't overwrite startup scripts, Cloudflare configuration, secrets, or unrelated project files in a narrow patch.

## Project handoff and regeneration convention

This README is the **single cumulative development record and chat-handoff document**, following the NewsGo, sailing, and stocks project workflow. Update it **every time source code is regenerated**. Deliver a ZIP patch containing only changed source/test files and the complete revised root `README.md` (do not create version-specific `SCRAPERGO-v*.md` notes). A documentation-only regeneration should contain `README.md` alone.

On a new ChatGPT conversation, supply this README and, when local changes exist, the current source archive or affected files. Treat local verified code as authoritative over older GitHub commits. Establish the version, tests, live behavior, unresolved issues, and next action before generating patches. Keep semantic version changes tied to deliberate release preparation rather than every minor iteration.

**Handoff state (October 19, 2026):** The user reports that local `main` is up to date with `origin/main` and the working tree is clean. GitHub `main.go` declares `appVersion = "0.2.49"`, and its latest verified commit is the v0.2.49 hyperlink-preservation release. This documentation update is prepared locally for the next commit; it has not been pushed or tested on the user's Mac. No pending feature request or known new regression was supplied at this handoff.
