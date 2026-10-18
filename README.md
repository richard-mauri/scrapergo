# ScraperGo v0.2.17

ScraperGo is a lightweight, Go-based **clean web reader**. It accepts HTTP(S) links from **any website**, retrieves accessible HTML, and uses Go Readability to identify the main content and present text and images without the surrounding page clutter. You can read blog posts, essays, tutorials, reference articles, recipes, travel writing, reviews, and news. Google News link resolution is an additional feature, **not a requirement**.

ScraperGo is designed for **reading now**, not for bookmarking or archiving. It works best with text-oriented web pages; interactive applications, catalogs, dashboards, login-only pages, and heavily scripted sites may not extract cleanly. It does not bypass subscriptions, registration requirements, or access controls.

**Choose the setup that describes your machine:**

- **Server host (Cloudflare Tunnel):** A Mac runs ScraperGo on `localhost:8083`, with an existing `cloudflared` tunnel and hostname route exposing it as `https://scrapergo.themauris.org/`.
- **Remote Safari client:** Another Mac runs **only** Safari and the temporary ScraperGo Web Extension. It connects to `https://scrapergo.themauris.org/`. It does not need Go, the ScraperGo executable, or `cloudflared`.
- **Local-only Safari client (optional):** Safari and ScraperGo run on the same Mac, and the extension points to `http://localhost:8083/`. No tunnel required.

## A. Server host: ScraperGo + Cloudflare Tunnel

### 1. Download and start the server

Download the appropriate binary from the [v0.2.17 release](https://github.com/richard-mauri/scrapergo/releases/tag/v0.2.17): `scrapergo-darwin-amd64` for Intel Macs or `scrapergo-darwin-arm64` for Apple Silicon. The release includes `SHA256SUMS` for verification.

On an Intel Mac, from the directory containing the executable:

```bash
chmod +x scrapergo-darwin-amd64
./scrapergo-darwin-amd64 -port 8083
```

On Apple Silicon, substitute `scrapergo-darwin-arm64`. An unsigned downloaded binary may require approval through macOS Privacy & Security before it can run.

Verify the **local** reader first by opening `http://localhost:8083/` in Safari. Keep the server process running.

### 2. Configure the existing Cloudflare Tunnel

Configure your existing **Cloudflare Tunnel public hostname route** as follows:

| Setting | Value |
|---|---|
| Public hostname | `scrapergo.themauris.org` |
| Service type | `HTTP` |
| Service URL | `http://localhost:8083` |

These settings belong to the Cloudflare Tunnel whose `cloudflared` connector runs on the **server Mac**. If you configure the tunnel through Cloudflare's dashboard, use its published-application/public-hostname route settings. If you manage ingress through a local configuration file instead, ensure the corresponding hostname routes to `http://localhost:8083`; don't mix two conflicting configuration methods.

Keep both ScraperGo and that tunnel connector running. From a different device, open `https://scrapergo.themauris.org/` and confirm the ScraperGo page appears. HTTPS terminates at Cloudflare; the local service still uses HTTP on port 8083.

**Security:** Exposing the reader publicly allows outsiders to reach its URL-fetch endpoint unless you protect it. Use appropriate Cloudflare Access/authentication and rate controls before offering the service beyond trusted clients. A Cloudflare Tunnel alone is not authentication. Any access policy must also work for Safari's navigation to the reader.

### 3. Optional: build from source on the host

With Go installed and the repository checked out:

```bash
go mod tidy
go test ./...
./devbuild.sh
./dist/scrapergo-darwin-amd64 -port 8083
```

On Apple Silicon, run the `arm64` binary. `./devbuild.sh` builds both macOS architectures and `dist/SHA256SUMS`. You do **not** need to build from source to use the release binaries.

## B. Remote Safari client (no local ScraperGo server)

On each client Mac, download the [repository source ZIP](https://github.com/richard-mauri/scrapergo/archive/refs/tags/v0.2.17.zip) and extract it. The required files are in `safari-extension/`. **The three release assets contain only the server binaries and checksums; they do not currently bundle the extension.**

Follow [Safari extension installation](safari-extension/README.md), and set its **Reader URL** to:

```text
https://scrapergo.themauris.org/
```

No Go toolchain, server binary, tunnel configuration, or local port 8083 service is required on a remote client.

## C. Optional local-only Safari client

Run the server on the same Mac at `http://localhost:8083/` and follow the same [Safari extension installation](safari-extension/README.md), but set **Reader URL** to:

```text
http://localhost:8083/
```

This is separate from the Cloudflare-hosted configuration.

## Using the Safari extension

On **any website**, right-click a link to a page you want to read (for example a blog post, tutorial, review, recipe, essay, or news article) and choose **Open Link in ScraperGo**. The extension navigates directly to ScraperGo without first opening the linked website in a visible intermediate tab. It reuses an existing ScraperGo tab, bringing it forward; when none exists, it opens one. Extraction happens on the ScraperGo server. For Google News links, the server also attempts to resolve the encoded news URL to its source website.

If a website refuses HTTP retrieval, the reader may report an HTTP status; an HTTP 403 does not necessarily mean a paywall. Some sites provide only previews, require registration, depend on JavaScript, or block externally loaded images. ScraperGo does not bypass access restrictions. Complex, non-article layouts may yield incomplete or unsuitable reading views.

## Source development and publishing

From the repository:

```bash
go mod tidy
go test ./...
./devbuild.sh
```

To launch an Intel development binary:

```bash
./dist/scrapergo-darwin-amd64 -port 8083
```

After committing and pushing changes to `main`, `./devbuild.sh --release vX.Y.Z` can publish both macOS binaries and checksums using an authenticated GitHub CLI. It requires a clean working tree and will reject an existing tag. **Do not reuse the already published v0.2.17 tag for documentation changes.**

The temporary Safari extension has its own lifecycle: reload it after JavaScript or manifest changes, independently of rebuilding the Go server.

## Files and local configuration

`main.go` handles HTTP retrieval, Google News resolution, and the API; `readability.go` integrates Go Readability; `ui.go` implements the reader page; `safari-extension/` contains the Safari Web Extension; `devbuild.sh` builds and publishes macOS binaries.

Do not package or publish `.cloudflared/`, tunnel credentials, local secrets, or a developer's existing `.gitignore`. These README files are documentation-only replacements.
