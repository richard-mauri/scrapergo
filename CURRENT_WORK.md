# Current work

**Authoritative baseline:** GitHub `main` at `4c169d9f73cd1c303aeee10f79b9d649d34470e6` when this branch was created (2026-11-06). Re-check branch and PR state at next session. No open PRs were reported immediately before this task; the older documentation PR #105 reference was stale.

**Active change:** [Issue #106](https://github.com/richard-mauri/scrapergo/issues/106), branch `fix/live-update-boundaries`, a focused runtime fix targeting `main`. The user approved working on live-feed article boundaries, **not** merging or releasing.

**Observed:** CBS News Nov 6 live coverage has successive legitimate editorial updates. The adviser-to-Iran-leader paragraph belongs to the earlier update, ahead of a tanker update; it must not be deleted as a teaser. ScraperGo's flat typography made the update boundary hard to distinguish.

**Implementation on branch:** `internal/extraction/live_headings.go` restores `h3` kind only when an existing paragraph exactly matches a unique source `h2/h3/h4` inside an editorial `article/main` scope on a `/live/` URL. The Readability, legacy and partial-extraction paths invoke it. `internal/ui/static/reader.js` scopes live-feed presentation to the URL path, and `reader.css` adds stronger h3 typography and separators. Synthetic extraction and UI tests added. Ordinary article layout and all legitimate prose are intended to remain unchanged.

**Limitations/risks:** Source HTML from the specific live CBS page and the structured `/api/read` response have not been captured; confirm that the original update headings are h2/h3/h4 in the fetched HTML and that Readability preserved matching text blocks. Timestamp and byline recovery are **not** implemented by this branch; they require source-backed capture and follow-up work. No Go tests, build, CI, Mac, live publisher, Safari or print acceptance has yet been performed. The local execution environment cannot clone GitHub due to DNS network restrictions. Do not claim the regression tests pass without CI.

**Next actions:** Open/update a focused draft PR, check the exact-head Go CI (`go test -count=1 ./...` and `go build ./...`), inspect the PR diff, then owner-side Mac/Safari article and print acceptance. Preserve all existing recovery/deduplication and security boundaries. Update PR description with exact evidence and limitations before handoff. Obtain explicit owner approval before merge, tag, release or branch cleanup.

**Deferred:** Verify source-backed timestamp/byline behavior and the actual CBS feed HTML. No unrelated reader/theme or extraction refactor is in scope.
