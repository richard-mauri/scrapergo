# Current work

**Authoritative merged baseline:** GitHub `main` at `4c169d9f73cd1c303aeee10f79b9d649d34470e6` when this branch was created on 2026-11-06. Recheck GitHub's current PR/branch state when resuming.

**Current goal:** Diagnose the loss of publisher HTML update headers and timestamps in ScraperGo without changing production extraction. The user correctly identified that URL-specific detection and timestamp-tail parsing introduced in draft PR #107 are brittle; do **not** merge that production change before source-structure analysis.

**Active branches/PRs:** Issue [#106](https://github.com/richard-mauri/scrapergo/issues/106) tracks the live-update problem. Draft [PR #107](https://github.com/richard-mauri/scrapergo/pull/107) remains open on `fix/live-update-boundaries`, but its heuristic implementation is not recommended for merge. This separate branch, `test/readability-structure-probe`, was created from `main` to isolate diagnostics from the unresolved fix.

**Diagnostic implementation:** `cmd/readabilityprobe` reads a local HTML file and reports source DOM versus Readability `RenderHTML` heading/time/article/section structure plus `RenderText` output. Three synthetic fixtures under `testdata/readability` cover semantic update containers, generic-div update containers, and a normal article. Tests verify report generation and source inventory. This tool does not contact publishers, change production services, or require the ScraperGo UI.

**Next:** Confirm exact-head Go CI, use the documented commands in `cmd/readabilityprobe/README.md` on the owner's Mac, compare saved local output (source / RenderHTML / RenderText), and—only with permission and without committing protected material—test a local copy of a relevant publisher page. Determine where the document structure is lost before redesigning the production fix. Do not merge, tag, release or deploy without explicit owner approval. Mac/Safari publisher checks have not been done.

**Out of scope:** Broad extractor refactoring, changing fetcher security, changing suite configuration, modifying Safari extensions or UI CSS, adding publisher-specific URL rules, or removing legitimate editorial content.
