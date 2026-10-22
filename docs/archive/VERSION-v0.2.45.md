# ScraperGo v0.2.45 — single version source

`appVersion` is defined once in `main.go`. The HTTP page calls `renderedIndexHTML()` to substitute that value into the static HTML template, and the startup log uses the same constant. Header tests use the rendered page rather than stale hard-coded version strings.

This patch is based on the v0.2.44 inline-image recovery main.go. It does **not** modify image extraction. Overwrite only files in this patch; your Tanner image and launcher configuration are unaffected.
