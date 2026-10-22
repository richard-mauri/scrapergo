# ScraperGo v0.2.49 — Readability hyperlink preservation

Use Readability RenderHTML output to preserve ordered inline hyperlink parts on already selected paragraphs, including after structured-document recovery. Existing article selection, image placement, and hero deduplication paths are unchanged. This is a generic HTML semantic fix, not a publisher-specific rule.

Install over v0.2.48: unzip -o scrapergo-v0.2.49-readability-links.zip in the ScraperGo repository, run go test ./... and ./devbuild.sh, then restart ScraperGo. Confirm the three PDF links in the Sanders press release using the /api/read JSON and Safari.

The standalone semantic parser tests passed; complete project tests and live publisher HTTP integration must be run on the Mac.
