# ScraperGo v0.2.8 — Experimental Readability Integration

This candidate retains the Safari extension and Go UI from v0.2.7, and adds the independently maintained, MIT-licensed `codeberg.org/readeck/go-readability/v2` parsing library. It uses the library on HTML that ScraperGo already retrieved; it does **not** itself solve publisher HTTP 403 or 401 restrictions. The existing parser is a fallback.

**Status:** Source integration prepared; dependency compilation and live AP News retrieval must be tested on your Mac. Do not publish to Releases yet.

## Install and test

Run `go mod tidy` (first use requires internet), `go test ./...`, then `go run . --port 8083`. To build Intel and Apple Silicon executables run `./devbuild.sh`. For Safari, reload the currently installed extension after updating if necessary; its configuration has not changed.

Reader displays a cover image if Readability finds an absolute HTTP(S) image URL; article body is currently plain text, so inline images and clickable inline links are not yet preserved. Do not consider the image work complete.

The first comparison is AP News, Phys.org, The Independent, and Yahoo Finance. Record HTTP status, extraction method, word count, and image presence. The AP News 403 will likely persist unless retrieval changes; do not mislabel HTTP 403 as a paywall.

# ScraperGo — v0.2.2

A lightweight ad-free **Go** article reader. Paste an article URL into the web interface and extract the main editorial text while omitting much of the page furniture, such as ads, menus, promotions, sidebars, widgets, and newsletter calls to action. Download the result as `.txt` or `.md`, or print/save it as a PDF from your browser.

**Example URL:** https://247wallst.com/investing/2026/10/16/taiwan-is-planning-4-billion-industrial-parks-in-texas-for-nvidias-suppliers/

## Run on macOS or Linux

Requires Go 1.21 or later. **No third-party Go packages** and no Python, Node, Homebrew, or Docker needed.

```bash
go run .
```

Open http://localhost:8080, paste a URL, and click **Read article**. To create a native executable:

```bash
go build -o scrapergo .
./scrapergo --port 8081
```

Run tests: `go test ./...`.

## Deploy on Render

Create a **Web Service** from your GitHub repository with environment **Go**. Set build command to `go build -o app .`, start command to `./app`. Render supplies `PORT` automatically. Alternatively, `PORT=9000 go run .` selects a local port.

## How it works

The Go backend downloads the source HTML, looks for the most article-like container, selects paragraph and heading content, and filters typical ad/promotion and site-navigation patterns. The browser displays only extracted text, not embedded ads or trackers from the original site. All content is fetched server-side; the original page's JavaScript is never executed.

## Important limitations

* Uses HTML heuristics, not a universal readability model; some sites may include stray non-article content or omit article material. Validate extraction before relying on it.
* JavaScript-rendered or protected content might be unavailable. The application **does not** bypass paywalls, access controls, or login requirements.
* Downloads retain content attribution and source URL. Respect publishers' terms and copyright when saving or redistributing text.
* Public hosting provides an on-demand server-side URL fetcher. The backend restricts localhost, private/reserved IP addresses, URL credentials, and non-HTTP(S) schemes, and rechecks redirects and connection DNS. For an internet-facing production deployment, also add authentication and request rate limiting.
* This version extracts text, not article images, tables, or interactive charts. Some embedded article promotions are filtered with keyword heuristics.
* Some pages use HTML structures not handled perfectly by the standard-library-only parser.

## Files

`main.go` — HTTP server, extraction, request safety.

`ui.go` — responsive web interface and client-side text/Markdown downloads.

`main_test.go` — extraction and private-network protection tests.

`go.mod` — module definition.

## Changes in v0.2.2

Corrects article selection when surrounding page elements have promotional class names; adds JSON-LD articleBody fallback and a more informative extraction failure message. Live access to 24/7 Wall St. is not verified in this environment.

## Building macOS binaries

Run `./devbuild.sh` from the project directory (with Go installed). The script
runs the tests, then builds two CGO-disabled standalone executables in `dist/`:
`scrapergo-darwin-amd64` for Intel Macs and `scrapergo-darwin-arm64` for Apple
Silicon Macs (including M5). Pass `--port 8081` when running to change the port.

These binaries do not require the Go toolchain on destination machines. macOS
minimum-version compatibility is determined by the Go version used to build;
older macOS releases and PowerPC Macs are not universally supported. Because
the build is unsigned, macOS Gatekeeper may warn when transferring the binaries.

## Open an article from Safari (bookmarklet)

With ScraperGo running on port **8081**, you can send Safari's current page
straight to ScraperGo without copying the URL. This uses a regular Safari
bookmark; it does **not** require a browser extension.

1. In Safari, show the Favorites bar (**View → Show Favorites Bar**).
2. Create a bookmark for any page and name it **Read in ScraperGo**.
3. Edit the bookmark's **Address** (not its name), replacing it with the
   **entire line** below:

```javascript
javascript:(function(){location.href='http://localhost:8081/?url='+encodeURIComponent(location.href)})()
```

4. Start ScraperGo: `./dist/scrapergo-darwin-arm64 --port 8081`
   (or `go run . --port 8081` when developing; use the amd64 binary for Intel).
5. Open a publicly accessible news article in Safari and click **Read in ScraperGo**.
   A new page loads the ScraperGo reader and automatically extracts that article.

**Important:** The bookmarklet replaces the current Safari tab with ScraperGo.
If you prefer to preserve the original page, duplicate its tab first. The
bookmarklet assumes port **8081**. If you run ScraperGo on a different port,
change `8081` in the bookmark Address. Safari may strip the `javascript:` prefix
when pasted into the address bar; edit the bookmark's Address field instead.
If JavaScript bookmarks are disabled by browser policy, this method will not work.

Alternatively, open `http://localhost:8081/?url=` followed by the URL-encoded
article address. Normal manual URL entry also continues to work.

## Command-line options (v0.2.2)

`--port 8081` (or `-port 8081`) sets the listening port. `--help` or `-h` displays usage and examples. Without `--port`, ScraperGo reads `PORT` from the environment (for Render compatibility), otherwise defaults to 8080. CLI takes precedence over `PORT`. Invalid ports are rejected. If the selected port is occupied, the error displays retry commands, including `go run . --port 8081`.

```bash
go run . --port 8081
./dist/scrapergo-darwin-amd64 --port 8081
./dist/scrapergo-darwin-arm64 --help
```

## Google News support (v0.2.8)

ScraperGo now tries to resolve `news.google.com/read/...` and `/articles/...` links to the original publisher URL before extracting content. This works with the Safari right-click extension and manual entry. Google uses an undocumented resolution endpoint; this can change or fail due to consent pages, network restrictions, or Google redesigns. Failures are reported distinctly rather than as an empty article. Original publisher paywalls are not bypassed.

Launch ScraperGo on the port configured in the Safari extension (for example `go run . --port 8083`).

## Google News resolution (v0.2.8)

Right-click a headline in Google News using the Safari extension, or paste its
`news.google.com/read/...` URL. ScraperGo tries to resolve the publisher URL
before extracting the article. v0.2.8 corrects the Google `Fbv4je` RPC frame
parser and improves diagnostics for consent pages, blocked requests, and missing
publisher URLs. Google's private endpoint is undocumented and can change; live
resolution is not guaranteed. If the resolver fails, open the publisher article
in Safari and submit its direct address to ScraperGo.


## Safari extension v0.2.0 (bundled)

This release includes `safari-extension/`, a temporary Safari Web Extension. On Google News, the extension opens the link in a background tab and observes navigation to a publisher address for up to nine seconds. If no publisher address is exposed, it hands the original link to ScraperGo’s existing server-side resolver. Browser-assisted resolution is best effort; some Google News links never navigate directly to the publisher.

Safari → Settings → Developer → Add Temporary Extension: select the `safari-extension` folder, then enable it in Settings → Extensions. Uninstall the earlier temporary extension first to avoid duplicate context-menu commands. In Safari Extensions settings, open the extension settings/options to configure the reader endpoint (default: `https://scrapergo.themauris.org/`).

Right-click a Google News link and select **Open Link in ScraperGo**. The browser may briefly open a background tab during link resolution. If extraction fails, the reader shows **Open original article in Safari**. This fallback does not bypass publisher access controls.

**Security:** The Cloudflare hostname is publicly reachable unless protected. Restrict it with Cloudflare Access or authentication before general use. ScraperGo does not receive cookies or browser DOM content from the extension.

## v0.2.8 paywall detection and browser-assisted reading

The Safari extension v0.3.0 attempts to read visible article paragraphs from a temporary publisher tab and forwards them directly to the ScraperGo reader. When unavailable, ScraperGo falls back to server-side extraction. This does **not** bypass paywalls or access controls. Explicit subscription markers on the server produce a subscription-required notice; HTTP 403 alone is **not** classified as a paywall. Browser-derived text passes through the configured ScraperGo origin, which may be Cloudflare Tunnel: avoid sending private articles to an untrusted endpoint.

Reload or replace the temporary Safari extension after installing. The reader origin defaults to https://scrapergo.themauris.org/.

## v0.2.9 retrieval experiment

This development version sets ordinary browser-navigation request headers and uses a per-request cookie jar. It preserves the existing DNS/private-network checks, redirects limit, Readability extraction, and Safari extension. HTTP retrieval logs contain request host, status, final host, content-type and elapsed time (not article bodies or cookie values). These changes are experimental: they do **not** guarantee AP News or other blocking publishers will accept a Go HTTP request. If a publisher still returns 403, use the existing Safari-assisted extraction method. No browser fingerprint spoofing, paywall bypass, or proxy service has been added.

Run `go mod tidy && go test ./... && go run . --port 8083` from the project folder. Test the AP News article against Readeck, and inspect terminal logs for `fetch host=apnews.com`.


## v0.2.10 image experiment

This development version preserves the v0.2.9 HTTP retrieval implementation. For Go Readability extraction, it selects a hero image from Readability metadata, falling back to `og:image` / `twitter:image`. It also attempts to collect inline images and figure captions from Readability-cleaned HTML, displayed after the article paragraphs in this initial implementation. Images are linked directly from publisher servers; publishers may restrict hotlinking. Raw publisher HTML is never injected into the application. The standalone Safari fallback remains unchanged.

After extracting the files, run `go mod tidy`, `go test ./...`, and `go run . --port 8083`. This version has not been compiled in the packaging environment due to unavailable Go module download access.


## v0.2.11 image CSP fix

The browser Content Security Policy now includes `img-src 'self' https: data:;`
so Safari may load article photographs served from HTTPS publisher image hosts
such as `dims.apnews.com`. Script restrictions and HTTP retrieval behavior are
unchanged. Third-party image hotlink restrictions may still prevent some images
from loading. Reload the ScraperGo page after restarting the server.


## v0.2.12: Direct-open Safari extension

The Safari right-click action now opens ScraperGo immediately, passing the clicked
URL as `?url=`. It does not navigate a temporary Safari tab to the original
publisher, avoiding the ad-filled page flash. ScraperGo's existing Go backend
resolves Google News URLs when possible and extracts public publisher HTML.

**Tradeoff:** the previous automatic Safari DOM-capture path is no longer
initiated by the context-menu action. Articles that only succeeded via browser-
assisted capture may fail until a separate, nonintrusive fallback is designed.
The previous content script is retained for compatibility but is not called by
the direct-open action. No Go retrieval/image code was changed.

**Activation:** after replacing `safari-extension/background.js` and
`manifest.json`, reload/reinstall the Safari temporary extension from its
actual source folder so the new background worker is active. Merely rebuilding
the Go server does not update the Safari extension. For the page's version
label, rebuild/restart the Go server from v0.2.12 source on port 8083.

Commands: `go test ./...`, then `go build -o scrapergo .`, then
`./scrapergo --port 8083`. The Go Readability dependency must be present
(or downloadable) for tests and builds.

## v0.2.13 access classification

The API now distinguishes free registration, an explicitly paid subscription, and an otherwise restricted article. HTTP retrieval failures are not considered proof of a paywall. The Safari extension and article extraction/retrieval paths remain unchanged.


## v0.2.14 restriction diagnostics

This development version preserves direct Safari extension navigation, article extraction,
images, and existing classification logic. When a publisher page contains restriction
markers, the reader now displays the **resolved publisher hostname** and the **matched
detection marker** (for example `structured-paywall-metadata` or `free-reading-phrase`).
Terminal logs record the same fields and the received HTML byte count, without logging
article text or cookies. This can reveal why an Axios free-registration prompt was
classified as a generic access restriction.

Test using the same Axios Google News link. If the result remains `Access Restricted`,
report the `Publisher` and `Detection` fields. This build intentionally does not infer
a free registration wall merely from the Axios domain.


## v0.2.15 — Axios HTTP 403 comparison

The initial HTTP request remains unchanged. On HTTP 403 only, ScraperGo
makes one additional request with an independently implemented, internally
consistent Chromium/Windows-style HTTP header profile inspired by inspection
of Readeck's open-source HTTP client. The retry uses a fresh cookie jar and
does not invoke Safari or create intermediate publisher tabs. The application's
DNS safety checks, redirect limits, image handling, Google News resolution,
and Readability extraction are unchanged.

The terminal now logs `profile=existing` or `profile=readeck-like` with the HTTP
response status, so a successful second attempt can be identified. This does
**not** duplicate Readeck's custom TLS behavior and might not resolve Axios.
Run tests before `./devbuild.sh` and evaluate the same Axios URL as well as AP News.


## v0.2.17 — single reusable ScraperGo tab

The Safari extension now searches all browser tabs for an existing tab on the configured ScraperGo origin. When found, it navigates that tab to the selected article, activates it, and attempts to focus its window. Otherwise, it creates one new ScraperGo tab. Overlapping clicks are serialized. The Google News resolver and all Go retrieval, Readability, and image logic remain unchanged. Reload the Safari extension for the new background script to take effect.
