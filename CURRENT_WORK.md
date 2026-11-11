# CURRENT WORK — ScraperGo handoff (November 11, 2026)

## Authoritative repository state

**GitHub `main` is authoritative.** Owner explicitly authorized merging [PR #120](https://github.com/richard-mauri/scrapergo/pull/120), closing Issue #119, deleting the completed branch, and returning the Mac checkout to `main` on November 11, 2026. PR #120 **merged successfully** as commit `038b275733be67094a061afa6dbb0bb1d4245357` (GitHub confirmed that commit at `main`); [Issue #119](https://github.com/richard-mauri/scrapergo/issues/119) was **closed as completed**. Follow-on work belongs to [new Issue #124](https://github.com/richard-mauri/scrapergo/issues/124) and will require **its own new branch and draft PR from the updated `main`**. Branch `fix/119-caption-duplication` has no remaining open PR and is approved for cleanup; **remote deletion and owner Mac checkout are still pending unless separately verified**. Do not imply that local files or remote branch have been deleted by an unavailable connector action. No release/tag/production deploy was requested.

The last merge includes generalized duplicate-text improvements, the New Scientist heading-anchored prose correction, source-backed guarded image-caption text reconciliation, diagnostic tools and tests. Previous open-PR `CURRENT_WORK.md` historical status is superseded by this snapshot. Before future work, read this file and `AGENTS.md`, `ARCHITECTURE.md`, `DEVELOPMENT.md`, inspect actual GitHub `main` and the relevant Issue/PR, and follow the full owner gate.

## Next priority — Issue #124: media-caption association and repeated hero caption

The owner's original-publisher and current-reader Safari screenshots of the November 2026 Ars Technica M6 MacBook Pro review establish **partial improvement, not perfect media presentation**. The owner Mac frozen HTML at `/tmp/pr120-ars-m6.html` (SHA256 `8bc65dad6714447409e5a6a3b6367339b0363f86f0703d4e6e500eb0a8055b9a`, **private, not in GitHub**) initially produced **68 canonical blocks, 1 exact pair 4/5 and 3 overlapping pairs 17/19, 22/24, 25/27**. After PR #120's guarded selected-Readability flattened-caption fix, the same snapshot produced **64 blocks, 1 exact pair and zero overlapping pairs**, with the three overlaps removed at `caption-reconciliation`. Original Safari Command-F showed each of the three captions exactly once. Owner's current ScraperGo screenshots show the redundant combined caption/credit paragraphs are gone, and each remaining photographer credit appears once.

**Still broken:** (1) Hero caption `Apple's M6 MacBook Pro. Credit: Andrew Cunningham` appears **twice** consecutively under hero photo. It is a plausible but **not source-proven** identity for remaining exact canonical pair **4/5** (digest `22f6070e195b`); the minimal source and selected HTML matcher returned **zero matches**. (2) The first paired Apple-logo and keyboard images appear, followed by **all four** captions/credits, **then the pair of ports images**, so the two ports captions precede their own images and photo-specific left/right associations are lost. Original publisher layout associates each photo with its caption and credit. A change that merely reduces duplicate text without preserving image association is insufficient.

Issue #124 owns investigation of `PositionArticleFigures`, image/gallery ordering, `internal/article` canonical block/group structure, `internal/reader`, and `internal/ui` rendering. Preserve source image-caption-credit associations, pair identity, source order and links instead of guessing from flattened text. Investigate 4/5 via bounded source and selected HTML serialization separately; **do not apply global paragraph dedup**. Protect true repeated prose and same-caption different photographs. The **four-block reduction for three flagged pairs** requires explicit content preservation review; a candidate pair count is not sufficient to prove losslessness.

## Unfinished acceptance and retained engineering evidence

- Earlier New Scientist paleo saved snapshot: **92 blocks / 15 duplicates → 77 blocks / 0 duplicates** after heading-anchored recovery correction. Other publisher regressions still require verification at the now-merged version.
- Prior 100-URL NewsGo-to-ScraperGo cohort: **100/100 successful reads; 54 screening candidates across 21 articles**, down from 100/27 before the prior correction. **This is not a final 100-URL rerun after the last Ars caption change**; content preservation is still unverified.
- Exact PR #120 final branch head `0a75b14ca0dfb6f5ab7da38119c955fdf5f0c0a9` passed GitHub [Go CI 43092889211](https://github.com/richard-mauri/scrapergo/actions/runs/43092889211) (`go test -count=1 ./...`, `go build ./...`). This does **not** establish Mac Safari/Print/PDF or live publisher correctness.
- A future structure-preserving HTML extraction experiment may be useful, but no replacement of Readability is authorized. Preserve the working improvements and evaluate experiments against them with source-backed tests before an architecture change.
- [Issue #123](https://github.com/richard-mauri/scrapergo/issues/123) documents a deferred opt-in staging/test agent cycle; [Issue #116](https://github.com/richard-mauri/scrapergo/issues/116) covers deferred shared-DOM/performance work. Do not combine them with Issue #124.

## Mac checkout and cleanup after merged PR #120

Before changing branches verify no uncommitted local work is present. When clean:

```bash
cd ~/projects/scrapergo
git status --short --branch
git fetch origin
git checkout main
git pull --ff-only origin main
git branch -d fix/119-caption-duplication
git fetch --prune origin
git status --short --branch
```

The owner also authorized deleting **the merged remote branch**; when no PR uses it, `git push origin --delete fix/119-caption-duplication` removes it. This is a *pending owner-side operation* if GitHub's Delete branch button has not been used; never claim connector deletion occurred without verifying. Do not reset local changes, force-delete a branch, delete `main`, or erase private files.

When beginning a new Issue #124 PR, start from the actual current `main`, use a new scoped branch and draft PR, and preserve source/selected HTML evidence with synthetic positive and negative tests. After CI at the **exact new PR head**, the owner Mac code/runtime sequence is `git fetch origin`, `git checkout <actual-new-branch>`, `git pull --ff-only origin <actual-new-branch>`, `go test -count=1 ./...`, `./devbuild.sh`, `./scrapergo.sh`; owner verifies article content, both photos in each row, each nearby caption/credit, links, Safari and Print/PDF, and reruns saved Ars/New Scientist and the 100 URL cohort. CI does not replace browser acceptance. Merge/tag/release/deploy only with explicit owner authorization.

## Security and preservation constraints

ScraperGo is a Go service with canonical model in `internal/article`, extraction in `internal/extraction`, reader in `internal/reader`, media identity in `internal/media`, source parsing in `internal/sourcehtml`, safe retrieval in `internal/fetcher`, and presentation in `internal/ui`. Do not bypass URL validation, network/redirect protections, access-policy boundaries, image delivery safeguards, or package ownership in `ARCHITECTURE.md`. Private owner Mac files `/tmp/pr120-ars-m6.html`, `/tmp/pr120-paleo.html`, `diagnostics/recovery.jsonl`, and `.recovery-key` are not GitHub assets: do not upload, delete, reset, or expose them. The native ARM64 `./scrapergo.sh` serves on port 8083 with existing narrow Cloudflare tunnel and private token; do not change tokens, launcher/build counter, tunnel configuration, other apps, or semantic version without a separate explicit request.
