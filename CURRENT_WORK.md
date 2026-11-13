# CURRENT WORK — ScraperGo (November 12, 2026)


## November 13, 2026 — temporary guest authentication in existing PR

- **Coordination:** Homepage Issue #32 / draft PR #33 is the sole feature narrative. ScraperGo draft PR #133 remains the one companion PR; no new issue or PR, no unrelated services, no merge/restart.
- **Decision:** Guest `guest-xxx` IDs are **temporary authenticated Homepage identities** (one hour), not unauthenticated ScraperGo operating mode. Homepage issues, verifies, expires, and garbage-collects them entirely in process memory. ScraperGo continues to consume its existing Homepage child-handoff/profile authentication protocol without adopting a guest-token store.
- **Browser behavior:** The ScraperGo URL input and Read button are disabled until Homepage verifies identity. A user-visible modal explains how to visit Homepage and sign in with a named credential or one-hour temporary guest; it only opens Homepage on a user's explicit click. Never redirect or reload automatically. Once validated, it shows `Signed in as richard` or `Signed in as guest-xxx` on the same header row.
- **Persistence:** Temporary guests may use the reader but cannot save or restore server preferences. Homepage enforces refusal on storage endpoints; ScraperGo does not gain a separate settings store.
- **Validation:** CI tests verify sign-in prompt text, disabled reader controls, no automatic navigation and existing identity recovery. Owner Mac requires `darwin-arm64` and existing `./scrapergo.sh` launcher, only after explicit deployment approval. Safari needs named and guest identity, inactivity, expiration, and cross-service activity verification.

## Current scope — Homepage activity identity integration

- **Authoritative baseline:** GitHub `main` contains merged initial activity reporter PR #132. Existing independent diagnostics work remains out of scope.
- **Branch:** `feat/132-suite-presence-identity` from current `main`; companion to Homepage draft PR #33 and coordinating Homepage Issue #32. No new ScraperGo login requirement.
- **Change:** ScraperGo remains publicly readable as a guest. When arriving with a verified Homepage-issued child handoff, the browser stores it only in origin-scoped sessionStorage and removes the fragment; the reporter verifies the bearer against Homepage's profile endpoint on each presence ping. The validation returns only origin-scoped telemetry aliases for the parent session and user; ScraperGo deduplicates tabs sharing a session and sends aggregate counts, never identities or credentials, to Homepage.
- **Inline presentation refinement:** Owner requested the same suite navigation style shown in MauriStocks: ScraperGo's Back to Homepage link and verified sign-in/guest state now share one horizontal row, with teal bold status. Added UI regression to ensure the status is not rendered as a separate paragraph.
- **Visible session indicator:** The reader now displays **Guest — not signed in**, **Signed in as <user>** after server validation, or **Unable to verify sign-in** for validation errors. A new read-only `/api/activity/identity` endpoint serves a same-origin status response without revealing raw tokens or aliases. Identity checks are refreshed while visible; article access remains public. Regression tests cover guest/signed-in/unavailable responses and UI rendering.
- **Homepage dependency:** PR #33 adds the allowlisted `/launch/scrapergo` bridge, authenticated Homepage app launch, scoped profile validation, and pseudonymous aliases. ScraperGo profile access is not granted account-storage privileges. The NewsGo → ScraperGo link/recovery path is **not yet implemented** and requires coordination before final acceptance.
- **Tests/risks:** New reporter regression checks same-session deduplication, independent sessions, shared user and guest counting. Need exact-head Go CI, completed auth revocation/invalid-handoff, fragment contract, and temporary Homepage validation outage tests; full end-to-end Safari tests, and owner Mac/Safari acceptance. The handoff is advisory for telemetry and must never gate article access or relax fetcher/SSRF rules.
- **Mac acceptance when ready:** `cd ~/projects/scrapergo`; `git fetch origin`; `git checkout feat/132-suite-presence-identity`; `git pull --ff-only origin feat/132-suite-presence-identity`; `go test -count=1 ./...`; `./devbuild.sh`; `./dist/scrapergo-darwin-arm64 --version`; `./scrapergo.sh`. The last command restarts production ScraperGo and its tunnel and must run only if Richard explicitly approves deployment. Test direct anonymous reading and Homepage handoff in Safari after both companion builds.
- **Mac documentation correction:** `DEVELOPMENT.md` now matches `main`'s Apple Silicon `scrapergo.sh` launcher, using `dist/scrapergo-darwin-arm64` and `.activity.env`; the Intel artifact is compatibility output only. No runtime launcher change.
- **Restrictions:** no merge/tag/version bump/release/restart, production token changes, root extractor modifications or unrelated PR changes.

## Earlier development notes (verify before acting)

## Active Issue #131 — Homepage suite activity reporter

**Current scoped implementation:** [Issue #131](https://github.com/richard-mauri/scrapergo/issues/131), branch `feat/131-suite-activity` from merged GitHub `main` (baseline `92f68f16f846dbf3ad247a243ff3d8095c6636ed`), [draft PR #132](https://github.com/richard-mauri/scrapergo/pull/132). This is independent of *still-draft* diagnostic PR #126 on `experiment/125-structure-preserving-extraction`; do not merge, modify or replace that diagnostics branch as part of telemetry.

- **Implemented:** optional `SCRAPERGO_ACTIVITY_URL` / `HOMEPAGE_ACTIVITY_KEY_SCRAPERGO` server-only reporter to Homepage's `scrapergo` identity; ephemeral anonymous/unverified same-origin browser heartbeat; rolling 2/5/15/60-minute counts; active API request work (read/classify/PDF/thumbnail/image/diagnostics); 60-second private aggregate reports with nonce and timestamp. Code confined to `internal/activity`, `internal/app`, and `internal/ui`. No production root/extraction/fetcher/SSRF modifications.
- **Privacy:** no publisher URL, article text, individual accounts, tab IDs or browser/session secrets exported to Homepage. Signed-in browser classification is unverified and included under anonymous. Missing reports are unknown, never a restart permission.
- **Verification:** CI #684 **passed** tests, build and existing diagnostic checks on code checkpoint `bf89d76baf97927011171d4c0691cb8f2c203764`. A newer loopback-only test listener override and regression test were added to avoid exposing the local test service on all interfaces; final-head CI must be checked. Mac/Safari test and actual Homepage-to-ScraperGo integration remain **unverified**. No merge/deploy/release/tag or service restart.
- **Next:** verify final-head CI. Isolated owner Mac test on port **8088** with `SCRAPERGO_ACTIVITY_TEST_LISTEN=127.0.0.1:8088` against disposable Homepage test on **8788**, after checking/stopping only a previous *test* listener on 8788. Never stop production Homepage 8787 or ScraperGo 8083. Use a new private per-service key and existing disposable Homepage admin/history. After Safari acceptance, owner decides whether to merge, in what order, and how to configure private keys for controlled rollout.

## Owner Mac test/build workflow (the launch command below needs explicit production restart approval)

```sh
cd ~/projects/scrapergo
git fetch origin
git checkout feat/131-suite-activity
git pull --ff-only origin feat/131-suite-activity
go test -count=1 ./...
./devbuild.sh
./scrapergo.sh
```

**Do not run `./scrapergo.sh` for isolated acceptance or without owner authorization.** It owns a production process and tunnel. For the smoke test use a detached test worktree, `go build -o` to a fresh temp directory, and `--port 8088` only. Preserve local build counters, credentials, live diagnostics and unpushed edits.

---

## Previous development context and independent diagnostics work

## Authoritative repository state

**GitHub `main` is authoritative.** Owner explicitly authorized merging [PR #120](https://github.com/richard-mauri/scrapergo/pull/120), closing Issue #119, deleting the completed branch, and returning the Mac checkout to `main` on November 11, 2026. PR #120 **merged successfully** as commit `038b275733be67094a061afa6dbb0bb1d4245357` (GitHub confirmed that commit at `main`); [Issue #119](https://github.com/richard-mauri/scrapergo/issues/119) was **closed as completed**. Follow-on work belongs to [new Issue #124](https://github.com/richard-mauri/scrapergo/issues/124) and will require **its own new branch and draft PR from the updated `main`**. Branch `fix/119-caption-duplication` has no remaining open PR and is approved for cleanup; **remote deletion and owner Mac checkout are still pending unless separately verified**. Do not imply that local files or remote branch have been deleted by an unavailable connector action. No release/tag/production deploy was requested.

The last merge includes generalized duplicate-text improvements, the New Scientist heading-anchored prose correction, source-backed guarded image-caption text reconciliation, diagnostic tools and tests. Previous open-PR `CURRENT_WORK.md` historical status is superseded by this snapshot. Before future work, read this file and `AGENTS.md`, `ARCHITECTURE.md`, `DEVELOPMENT.md`, inspect actual GitHub `main` and the relevant Issue/PR, and follow the full owner gate.

## Active draft PR #126 — retain offline diagnostic tools only

**Decision (November 12, 2026):** [Issue #125](https://github.com/richard-mauri/scrapergo/issues/125) / [draft PR #126](https://github.com/richard-mauri/scrapergo/pull/126), branch `experiment/125-structure-preserving-extraction`, is narrowed to retaining the **new, previously unmerged diagnostics and their unit tests only**. GitHub `main` remains authoritative; the PR is open, draft and **unmerged**. Its change set adds `cmd/structureprobe`, `cmd/selectedstructure`, and the supporting `internal/sourcepreserve` package/tests. No existing production extractor, reader, UI, model, fetcher, or runtime setting changes. Prior proposals to implement a replacement structural extractor, change image eligibility or caption rules, or continue wrapper-heuristic development are **abandoned**. Experimental semantic conversion remains a *diagnostic representation*, not a candidate production path. PR #130 / Issue #129 are closed without merge; the proposed configurable first-12 limit is not included.

**How to prepare the owner's Mac:** start in `~/projects/scrapergo` with a clean working tree; do not discard local edits. Check out the existing branch and run the required code-PR validation (normal `devbuild.sh`, **not** `--release`):

```bash
cd ~/projects/scrapergo
git status --short --branch
git fetch origin
git checkout experiment/125-structure-preserving-extraction
git pull --ff-only origin experiment/125-structure-preserving-extraction
go test -count=1 ./...
./devbuild.sh
./scrapergo.sh
```

The normal build runs tests, the benchmark smoke check, and builds both Mac architectures; a separate `go build ./...` is unnecessary in this Mac sequence. Launch/Safari at `http://localhost:8083` verifies ordinary reader, images/captions and Print/PDF were not disturbed, **not** that offline diagnostics work. The commands below run without the server, a restart, a live publisher fetch or any private credentials. They accept a locally saved HTML file and its original HTTP(S) article URL, which resolves relative paths; `-url` is **not fetched**. On the Mac the existing private Ars snapshot is `/tmp/pr120-ars-m6.html`; keep it private and do not commit/upload it.

### Tool 1 — `cmd/structureprobe`: inspect original-source image structure

**Purpose:** inspect the publisher's saved original HTML, requiring exactly one explicit `<article>` scope. The ordinary report inventories explicit figures/captions/credits, headings and paragraphs, safe links, unwrapped images, exact-URL repetitions and *tentative* compact text-sibling / bounded multi-image wrapper shapes. The latter are historical **observations only**, not reliable caption/gallery associations or production selection rules.

**Full inspection** (may print publisher prose, resolved URLs, captions and credits; keep output local):

```bash
go run ./cmd/structureprobe \
  -input /tmp/pr120-ars-m6.html \
  -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/'
```

Inspect `status`, `items`, `ambiguousFigures`, `unwrappedImages`, `uniqueUnwrappedImageURLs`, `repeatedUnwrappedImageOccurrences`, and the wrapper/group counters. `status: inconclusive` means zero or multiple explicit article scopes (not proof that an article has zero images).

**Content-free structural trace** (preferred for sharing diagnostic counts):

```bash
go run ./cmd/structureprobe \
  -input /tmp/pr120-ars-m6.html \
  -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/' \
  -structure-trace
```

Inspect `images[].imageIndex`, `pattern` (`unwrapped`, `compact-text-sibling`, or `bounded-multi-image-group`), `parentTag`, group counts and `ancestors[]` (up to six enclosing levels, before/after text-bearing sibling branches, explicit caption-element counts). This trace intentionally **skips images inside explicit `figure` elements**, and reports structural shapes rather than editorial image identity or guaranteed caption ownership. It does not output article text, image URLs, CSS classes or IDs.

### Tool 2 — `cmd/selectedstructure`: compare source, Readability and final reader images

**Purpose:** on the *same* locally saved HTML, compare three views: (A) the existing production extraction + reader's final canonical image blocks, (B) Readability's selected HTML parsed into a nonrendering semantic tree (headings, prose, links, lists/tables and explicit figure/caption relationships), and (C) the original source's one explicit `article` subtree. The command applies production-equivalent embedded-media stripping before comparison; it also observes the already-existing extraction and reader stage counters. It does **not** replace Readability or render the experimental tree.

```bash
go run ./cmd/selectedstructure \
  -input /tmp/pr120-ars-m6.html \
  -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/'
```

The JSON includes `selectedHTMLKinds` (semantic node-kind counts), `selectedHTMLImages`, `flatTextCharacters`, `readabilityTitlePresent`, `sourceSelectedImageAlignment` (source vs selected exact resolved image-URL occurrence matches and 1-based source/selected positions), `baselineImageAlignment` (final canonical image blocks vs selected positions, preserving a separate `canonicalLeadImagePresent` flag), and `baselineImageStages` (reached recovery stages with `beforeImages`/`afterImages`; especially `position-figures` and `reader-postprocessing`). `sourceStatus: inconclusive` means the original source cannot be uniquely scoped to one explicit article; it does not invalidate the selected/canonical comparison. JSON prints aggregate counts, index matches, stage names and status, **not article text or image URLs**.

**Interpreting the existing Ars evidence:** the owner's previous run observed 20 source and 20 selected image nodes, 19 exact source/selected URL matches (positions 1–19), and 13 final canonical image blocks matching selected positions 1–12 and 19. Extraction's `position-figures` stage changed 0 → 12 images, and reader postprocessing changed 12 → 13. The standalone lead `Article.ImageURL` is separate from those 13 blocks. Unmatched selected positions 13–18 and 20 are **not proven to be seven omitted editorial photographs**; responsive variants, noneditorial images and caption identity require local source evidence. A matching URL is not proof of visual equivalence. Stage *counts* cannot identify which individual photo was added or removed.

### Supporting package and limits

`internal/sourcepreserve` supplies the offline DOM inspections and structural traces (`inspect.go`, `trace.go`), the nonrendering Readability-selected semantic tree (`selected_html.go`), and ordinal exact-URL comparisons against source and final canonical blocks (`image_alignment.go`, `canonical_alignment.go`), with adjacent synthetic `*_test.go` files. It does not expose a production API and must remain out of the reader, extraction and UI pathways. The tools do not establish CSS-visible layout, identify responsive renditions as the same photograph, associate nonexplicit captions reliably, or certify any proposed extraction change.

**CI diagnostic smoke coverage (new in PR #126):** `.github/workflows/go-ci.yml` runs both new commands against the synthetic `testdata/diagnostics/offline-article.html` (one explicit source article, figures/captions and an unwrapped image). It executes `structureprobe` in full and `-structure-trace` modes, plus `selectedstructure`, then parses all three results as JSON and asserts key statuses and semantic/image fields. These are offline, network-free integration smoke checks, not proof of live-publisher or Mac/Safari fidelity. The existing Readability probe CI and Go test/build jobs are retained.

**Verified CI evidence:** the new diagnostics smoke step passed in [GitHub Actions run 43238001083](https://github.com/richard-mauri/scrapergo/actions/runs/43238001083) at code head `9bcf66ffd9c0607114500043708835ec90a5ba81`, alongside the Go tests/build and existing Readability probes. This documentation update creates a newer head whose exact-head CI must also be checked. Owner Mac tests/build/smoke/launch, diagnostic runs with private Ars HTML, and Safari regression checks remain unverified. Test on a second unrelated saved publisher when available. Do not merge, tag, publish or deploy without explicit owner authorization. Do not commit owner-private HTML, credentials or diagnostics.

## Next priority — Issue #124: media-caption association and repeated hero caption

The owner's original-publisher and current-reader Safari screenshots of the November 2026 Ars Technica M6 MacBook Pro review establish **partial improvement, not perfect media presentation**. The owner Mac frozen HTML at `/tmp/pr120-ars-m6.html` (SHA256 `8bc65dad6714447409e5a6a3b6367339b0363f86f0703d4e6e500eb0a8055b9a`, **private, not in GitHub**) initially produced **68 canonical blocks, 1 exact pair 4/5 and 3 overlapping pairs 17/19, 22/24, 25/27**. After PR #120's guarded selected-Readability flattened-caption fix, the same snapshot produced **64 blocks, 1 exact pair and zero overlapping pairs**, with the three overlaps removed at `caption-reconciliation`. Original Safari Command-F showed each of the three captions exactly once. Owner's current ScraperGo screenshots show the redundant combined caption/credit paragraphs are gone, and each remaining photographer credit appears once.

**Still broken:** (1) Hero caption `Apple's M6 MacBook Pro. Credit: Andrew Cunningham` appears **twice** consecutively under hero photo. It is a plausible but **not source-proven** identity for remaining exact canonical pair **4/5** (digest `22f6070e195b`); the minimal source and selected HTML matcher returned **zero matches**. (2) The first paired Apple-logo and keyboard images appear, followed by **all four** captions/credits, **then the pair of ports images**, so the two ports captions precede their own images and photo-specific left/right associations are lost. Original publisher layout associates each photo with its caption and credit. A change that merely reduces duplicate text without preserving image association is insufficient.

Issue #124 owns investigation of `PositionArticleFigures`, image/gallery ordering, `internal/article` canonical block/group structure, `internal/reader`, and `internal/ui` rendering. Preserve source image-caption-credit associations, pair identity, source order and links instead of guessing from flattened text. Investigate 4/5 via bounded source and selected HTML serialization separately; **do not apply global paragraph dedup**. Protect true repeated prose and same-caption different photographs. The **four-block reduction for three flagged pairs** requires explicit content preservation review; a candidate pair count is not sufficient to prove losslessness.

## Unfinished acceptance and retained engineering evidence

- Earlier New Scientist paleo saved snapshot: **92 blocks / 15 duplicates → 77 blocks / 0 duplicates** after heading-anchored recovery correction. Other publisher regressions still require verification at the now-merged version.
- Prior 100-URL NewsGo-to-ScraperGo cohort: **100/100 successful reads; 54 screening candidates across 21 articles**, down from 100/27 before the prior correction. **This is not a final 100-URL rerun after the last Ars caption change**; content preservation is still unverified.
- Exact PR #120 final branch head `0a75b14ca0dfb6f5ab7da38119c955fdf5f0c0a9` passed GitHub [Go CI 43092889211](https://github.com/richard-mauri/scrapergo/actions/runs/43092889211) (`go test -count=1 ./...`, `go build ./...`). This does **not** establish Mac Safari/Print/PDF or live publisher correctness.
- New offline structural diagnostics are retained in draft PR #126; the proposed production structural-extraction migration and wrapper-heuristic work are abandoned. Readability and current recovery remain unchanged.
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
