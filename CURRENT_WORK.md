# CURRENT WORK — ScraperGo handoff (November 11, 2026)

## Authoritative repository state

**GitHub `main` is authoritative.** Owner explicitly authorized merging [PR #120](https://github.com/richard-mauri/scrapergo/pull/120), closing Issue #119, deleting the completed branch, and returning the Mac checkout to `main` on November 11, 2026. PR #120 **merged successfully** as commit `038b275733be67094a061afa6dbb0bb1d4245357` (GitHub confirmed that commit at `main`); [Issue #119](https://github.com/richard-mauri/scrapergo/issues/119) was **closed as completed**. Follow-on work belongs to [new Issue #124](https://github.com/richard-mauri/scrapergo/issues/124) and will require **its own new branch and draft PR from the updated `main`**. Branch `fix/119-caption-duplication` has no remaining open PR and is approved for cleanup; **remote deletion and owner Mac checkout are still pending unless separately verified**. Do not imply that local files or remote branch have been deleted by an unavailable connector action. No release/tag/production deploy was requested.

The last merge includes generalized duplicate-text improvements, the New Scientist heading-anchored prose correction, source-backed guarded image-caption text reconciliation, diagnostic tools and tests. Previous open-PR `CURRENT_WORK.md` historical status is superseded by this snapshot. Before future work, read this file and `AGENTS.md`, `ARCHITECTURE.md`, `DEVELOPMENT.md`, inspect actual GitHub `main` and the relevant Issue/PR, and follow the full owner gate.

## Active Issue #125 — isolated source-DOM proof of concept (draft PR #126)

The owner directed development on [Issue #125](https://github.com/richard-mauri/scrapergo/issues/125). **Draft [PR #126](https://github.com/richard-mauri/scrapergo/pull/126)** uses branch `experiment/125-structure-preserving-extraction` from authoritative merged `main` baseline `2832ac7cde154af2910042cd030feb67f9f905d3`. Research plan: `docs/structure-preserving-extraction-125.md`. **No production extractor/reader/UI/fetcher changes**; no change to runtime default Readability, version, credentials, tunnels, or deployment.

**Current isolated implementation:** `internal/sourcepreserve/inspect.go` inspects an already-retrieved HTML string using `golang.org/x/net/html`; returns an `inconclusive` verdict for zero/multiple article nodes and otherwise reports ordered paragraph and heading text, safe HTTP(S) link URLs, unique-image `figure`/optional `figcaption` associations, explicit small/cite credits, skipped noneditorial regions, and ambiguous figure counts. It rejects scriptable markup, explicit hidden and dialog/landmark regions and narrow exact class/id noise tokens. It does not render arbitrary HTML or integrate into production. `cmd/structureprobe` is a local **offline, read-only** JSON report command. Synthetic tests cover paired figures/credit/links, repeated caption on distinct photos, legitimate repeated prose, ad/popup/hidden filtering, unsafe URLs, ambiguous figures and article scope. The report **contains source text/URLs** and should not be published with private publisher content. It is not yet a complete sanitizer or responsive gallery selector.

**Evidence:** Synthetic Go tests/build passed GitHub Actions at code commit `43c4bd51da8394e3010c0e3ff5c68d6072f875f3` ([run 43103373688](https://github.com/richard-mauri/scrapergo/actions/runs/43103373688)); later documentation/handoff commits need **exact-head** CI verification. No owner Mac local Ars/New Scientist comparison, full 100-URL cohort, benchmark, Safari or Print/PDF acceptance yet. CI does not imply production or publisher correctness.

**Next owner action after fetching this branch (if desired):** run `go test -count=1 ./...` and `go run ./cmd/structureprobe -input /tmp/pr120-ars-m6.html -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/'` locally; inspect JSON **privately** for `status`, `ambiguousFigures`, image/caption adjacency and unintended loss. The old `cmd/duplicateprobe -source-provenance` is the matched-source Readability reference. Because this prototype does not modify runtime code, neither build/launcher restart nor Safari acceptance is yet required; later code/runtime PR handoff must include full Mac checkout/test/build/launch and Safari/Print/PDF validation. Need enhance article selection, responsive variant disambiguation, nested semantic elements, correctness of filtering, and quantitative comparison before considering adoption. **PR remains draft, no merge/release/deploy without owner authorization.** Keep #124 focused on the existing media-caption defect; #116 shared parse performance and #123 agent/staging workflow remain separate.

## Latest owner Ars experiment — ambiguity breakdown next

The owner Mac ran **`go test -count=1 ./...` successfully** and replayed the unchanged private Ars M6 HTML on `cmd/structureprobe`. The first real publisher result is `status=candidate`, source-ordered headings/paragraphs and safe links, **one explicit `figure` with caption**, **19 ambiguous figure/image encounters**, **28 skipped noneditorial subtrees**. Those numbers do **not** prove complete ad removal or photo-caption recovery. The prototype captures much article prose but does not recognize the generic wrapper image groups from the triggering article. No private HTML or complete JSON content was committed.

**New on this same draft PR #126:** the offline-only `internal/sourcepreserve` report now provides `unwrappedImages`, `multiImageFigures`, `invalidFigureImages`, `multipleCaptions` in addition to total `ambiguousFigures`. It makes no media association decision. Synthetic tests validate classification and preserve independent prose beside unwrapped images. Code/test CI at commit `98241ce00f0b192351661f1325b671327092cee6` passed [GitHub Actions 43104059138](https://github.com/richard-mauri/scrapergo/actions/runs/43104059138); later docs/hand-off commits need exact-head verification.

**Next:** pull branch and rerun identical offline Ars capture; report only `status`, `ambiguousFigures`, `unwrappedImages`, `multiImageFigures`, `invalidFigureImages`, `multipleCaptions`, and `skippedNoneditorial`. This will show whether the next model should address general single-image wrappers or multi-image figures. Do not claim overall improvement over Readability, touch production extraction, merge/release/deploy or infer photo identities solely from those counts. Remain on draft PR #126; Issue #124 remains separate.

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
