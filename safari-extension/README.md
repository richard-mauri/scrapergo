# ScraperGo Safari right-click extension (prototype v0.1.0)

Adds **Open Link in ScraperGo** to Safari's context menu when right-clicking a link. It sends the selected URL to ScraperGo in a new Safari tab, without changing the original page.

## Requirements

- Safari version supporting temporary WebExtensions.
- ScraperGo **v0.2.3 or later** running locally at `http://localhost:8081`.

## Install temporarily (no Xcode required)

1. Unzip the extension ZIP.
2. In Safari, open **Safari → Settings → Advanced** and enable **Show features for web developers** (wording varies by Safari version).
3. In **Safari → Settings → Developer**, enable **Allow unsigned extensions**, if required.
4. In **Safari → Settings → Developer**, select **Add Temporary Extension…**, and select this extension folder containing `manifest.json`.
5. Enable the extension in **Safari → Settings → Extensions**, if not already enabled.
6. Start ScraperGo: `go run . --port 8081`, or run your compiled Mac binary with `--port 8081`.
7. Browse Google News, **right-click an article link** (not an empty page area), and choose **Open Link in ScraperGo**.

This extension is temporary: Safari removes it when Safari quits or after 24 hours. A permanently installed version needs a macOS Safari extension app package.

## Limitations

- Google News links can be encoded or redirect URLs. ScraperGo may not be able to extract some of these without a further resolver update. The extension passes the link that Safari exposes; it does not try to bypass article access restrictions.
- Port 8081 is hardcoded in `background.js`. If ScraperGo runs on a different port, change `SCRAPERGO_ORIGIN` there and reload the extension in Safari settings.
- This is an initial prototype, not yet tested interactively on macOS Safari. 
- The extension opens `http://localhost:8081/?url=<encoded URL>`; it does not transmit browsing data to any external extension service.
