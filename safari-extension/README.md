# ScraperGo Safari extension — v0.2.17 setup

This temporary Safari Web Extension adds **Open Link in ScraperGo** to the right-click menu for links. Use it on **any website** to send a link to ScraperGo’s clean reading view: blog posts, tutorials, recipes, essays, reviews, documentation articles, and news are all valid candidates. The extension **reuses one ScraperGo tab** instead of accumulating new tabs and does not first display the linked page in a separate publisher tab.

## Choose the Reader URL

| Configuration | Reader URL | What runs on this Mac? |
|---|---|---|
| **Remote Safari client** | `https://scrapergo.themauris.org/` | Safari + extension only |
| **Safari on server host via tunnel** | `https://scrapergo.themauris.org/` | Safari + extension; server and cloudflared already running |
| **Local-only Safari** | `http://localhost:8083/` | Safari + extension + local ScraperGo process |

The URL must point to an **already running and reachable ScraperGo service**. Remote Macs do not need Go, `cloudflared`, or the ScraperGo binary.

## Install the temporary extension (no Xcode)

1. Download the [v0.2.17 repository source ZIP](https://github.com/richard-mauri/scrapergo/archive/refs/tags/v0.2.17.zip) and extract it. Locate `safari-extension/` containing `manifest.json`, `background.js`, `content.js`, and the options files. The macOS binary release assets do not include these files.
2. In Safari, open **Settings → Advanced** and enable **Show features for web developers** (wording varies by Safari version).
3. In Safari's **Developer** settings/menu, allow unsigned extensions if your Safari version requires it.
4. Use **Add Temporary Extension…** in Safari's Developer settings/menu and select the `safari-extension` folder containing `manifest.json`.
5. In **Safari → Settings → Extensions**, enable **Open Link in ScraperGo**. If Safari asks for website permissions, review and grant the access required for the extension to operate on the pages whose links you want to open.
6. Open the extension's settings/options page. Set **Reader URL** to `https://scrapergo.themauris.org/` for a remote client, or `http://localhost:8083/` for a local-only setup, and save.
7. Confirm the chosen reader address opens normally in Safari. Then visit **any website**, right-click a link to a content-rich page, and choose **Open Link in ScraperGo**. Google News links also work when its URL resolver can identify the source.

Depending on Safari version and policies, the temporary extension may be removed on restart or expire; add it again when necessary. A permanently installed Safari extension would require separate macOS packaging.

## Expected behavior

The first right-click opens a ScraperGo tab. Subsequent right-clicks **reuse that tab** and bring it forward. If that tab is closed, the next link opens a new one. The extension passes the selected HTTP(S) URL to the configured reader, where the Go server extracts the main readable content. Google News URLs receive an extra resolution step. No intermediate source tab is opened.

## Update or troubleshoot

- **New tabs keep appearing:** Confirm the loaded extension is v0.2.17 and reload/remove-and-readd the temporary extension. Restarting ScraperGo alone will not update `background.js`.
- **The wrong server opens:** Revisit the extension's **Reader URL** options. The default is `https://scrapergo.themauris.org/`; it is not a hardcoded `localhost:8081` endpoint.
- **Connection refused:** On local-only installations, confirm ScraperGo is running on port 8083. On remote clients, verify the HTTPS server is reachable and the server host's tunnel is connected.
- **Page cannot be retrieved or cleaned:** The server may receive HTTP 403, limited preview content, or subscription metadata, or may be unable to extract an interactive/script-heavy page. The extension cannot bypass website access requirements.
- **Duplicate context-menu entries:** Disable or remove older temporary versions, then add only the current extension.

## Privacy and security

The extension sends the selected webpage URL to the configured ScraperGo origin. With the HTTPS tunnel configuration, the article URL is sent to `scrapergo.themauris.org` for server-side retrieval. Only use a reader endpoint you trust. Protect a publicly exposed reader with access controls as appropriate. This temporary extension is not a packaged or signed App Store extension.
