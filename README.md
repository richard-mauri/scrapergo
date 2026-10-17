# ScraperGo

**ScraperGo** is a lightweight, self-hosted article reader written in Go. Paste a publicly accessible article URL into its browser interface to extract the main editorial content and read it without the surrounding ads, navigation menus, promotions, and unrelated page clutter.

ScraperGo is designed for local use on macOS and can also be hosted as a Go web service.

## Features

- Clean, responsive browser-based reading interface.
- Article extraction with filtering for common advertising and promotional elements.
- Fallback extraction from structured article data (JSON-LD) when available.
- Download extracted content as plain text or Markdown.
- Print articles or save them as PDF using your browser.
- Configurable listening port with command-line options or the `PORT` environment variable.
- Go standard-library implementation, with no third-party Go dependencies.
- Build native executables for Intel and Apple Silicon Macs.

> **Limitations:** Extraction depends on a site's HTML and structured data. Some pages require JavaScript, authentication, or site-specific rules, and may not extract successfully. ScraperGo does not bypass paywalls or access restrictions. Respect publishers' terms and copyright when using or sharing extracted content.

## Requirements

For development from source, install a compatible Go toolchain. For a prebuilt macOS executable, Go is not required on the destination computer.

## Quick start

Clone the repository:

```bash
git clone https://github.com/richard-mauri/scrapergo.git
cd scrapergo
```

Run from source:

```bash
go run . --port 8081
```

Then visit **http://localhost:8081** in your browser, paste an article URL, and select the extraction action in the interface.

The default port is **8080** if no option or environment variable is supplied.

## Command-line options

```bash
go run . --help          # Display usage and examples
go run . --port 8081     # Listen on port 8081
go run .                 # Listen on port 8080 by default
PORT=9090 go run .      # Set port through an environment variable
```

The `--port` option takes precedence over `PORT`.

If the selected port is already in use, ScraperGo reports the conflict and suggests launching with a different port. For example:

```bash
go run . --port 8081
```

## Build for macOS

Use the included `devbuild.sh` script:

```bash
chmod +x devbuild.sh
./devbuild.sh
```

The build script runs the Go tests and produces two executables under `dist/`:

| Binary | Macs supported |
| --- | --- |
| `scrapergo-darwin-amd64` | Intel-based Macs |
| `scrapergo-darwin-arm64` | Apple Silicon Macs, including M1 through M5 |

Run the appropriate executable:

```bash
# Intel Mac
./dist/scrapergo-darwin-amd64 --port 8081

# Apple Silicon Mac
./dist/scrapergo-darwin-arm64 --port 8081
```

These binaries are compiled with `CGO_ENABLED=0`, eliminating third-party native-library dependencies. They are architecture-specific, not universal binaries. Compatibility still depends on the minimum macOS version supported by the Go compiler used to build them. Unsigned downloaded binaries may trigger macOS Gatekeeper warnings.

## Tests

```bash
go test ./...
```

## Hosting

ScraperGo uses an HTTP server and honors the `PORT` environment variable, making it suitable for hosting on platforms such as Render. A typical build command is:

```bash
go build -o scrapergo .
```

A typical start command is:

```bash
./scrapergo
```

The hosting provider can assign `PORT` as an environment variable. If you deploy publicly, review the security implications of allowing visitors to submit arbitrary URLs to a server-side fetcher.

## GitHub Releases

For public distribution, keep source files in the repository and attach compiled executables to versioned **GitHub Releases**. This avoids storing changing binary assets in Git history.

Recommended `.gitignore` entries:

```gitignore
/dist/
.DS_Store
```

For additional integrity checks, generate and publish SHA-256 checksums with releases.

## Project status

**Current documented version:** v0.2.2

ScraperGo is an early-stage project. Extraction quality varies across publishers, and websites can change their structure without notice.

## Repository

https://github.com/richard-mauri/scrapergo
