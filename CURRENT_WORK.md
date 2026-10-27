# Current active work — Issue #87 reader typography (2026-10-27)

- Issue: https://github.com/richard-mauri/scrapergo/issues/87; draft PR #88: https://github.com/richard-mauri/scrapergo/pull/88; branch: `fix/issue-87-sans-serif-reader` from merged main.
- Scope: CSS-only replacement of Georgia serif with system sans-serif for reading-view article title/body; retain current sizing, color, dark mode, table styles, exports and extraction. Regression test added in `internal/ui/reader_typography_test.go`.
- Validation: CI and owner's local Mac/Safari acceptance pending. No merge, semantic version, tag or release without explicit approval.
- Previous Issue #85 was squash-merged into main as `c80b4d8` and closed. Its prior notes below are historical, not active.

---

# Active Issue #85 — press-release institutional tail (2026-10-27)

- Issue: https://github.com/richard-mauri/scrapergo/issues/85; draft PR #86: https://github.com/richard-mauri/scrapergo/pull/86; branch `fix/issue-85-press-release-tail` from `main` (commit `23e6987`).
- Reproduction: US Mint press release source includes `About the United States Mint` after `Learn more`, before `Additional Resources`; ScraperGo omits this institutional section.
- Analysis: the Readability text pipeline and conservative anchored-tail recovery can omit one-paragraph supplementary sections; precise source DOM placement on the live Mint page not yet independently validated.
- Candidate: `internal/extraction/institutional_tail.go`, invoked after source editorial continuity; only appends one source-backed `About ...` heading and substantial paragraph sharing an editorial main/article ancestor with a retained substantial source anchor. Footer-exclusion and positive synthetic tests added.
- Pending: CI at latest PR head, source HTML / real-page verification, owner Mac/Safari checks, print and exports. No merge, tag or release authorization. Table fix #83 is already closed and on main.

---

# Active draft — Issue #83 table preservation (2026-10-27)

- Issue: https://github.com/richard-mauri/scrapergo/issues/83 ; draft PR: https://github.com/richard-mauri/scrapergo/pull/84 ; branch: `fix/issue-83-preserve-tables` → `main`.
- Source diagnosis: `internal/extraction/readability_pipeline.go` uses Readability `RenderText` and paragraphs, losing HTML table grid semantics; legacy extractor also flattened tables.
- Implemented: typed table rows/cells in `internal/article/model.go`; source-backed `TableBlock` and `RestoreSourceTables` plus fallback handling; browser semantic tables, responsive CSS and exports; synthetic extraction tests.
- CI: [Go CI #397](https://github.com/richard-mauri/scrapergo/actions/runs/40520565739) passed at commit `6535075c9c84a0267f8ab24632a19dd84998f4cb`. Any newer commit requires a new exact-head check.
- Still required: owner's Mac checkout/tests/build/restart, live US Mint reproduction, Safari/mobile viewport, print and Markdown/text export acceptance. Current matching requires >=3 separate cell text matches; fallback for a single merged paragraph not proven. Check exact source table location and false-positive injection risk before release.
- Keep draft. No semantic version, merger, release, NewsGo or unrelated active branch changes. Owner authorization required.

---

# Current active work — publisher HTTP status, Issue #81

- Base: GitHub main, independent branch `fix/issue-81-publisher-http-status`.
- Scope: preserve actual upstream HTTP status as optional `publisherHttpStatus` on /api/read error JSON and /api/classify classification JSON, and show the HTTP code in the reader's existing error text when available. The existing HTTP 422 API status remains unchanged.
- No changes to fetch behavior, classification decisions, successful caching, failed caching, request headers, NewsGo, or Safari coupling.
- Synthetic, offline read and classification tests added; CI and owner Mac/Safari verification pending.
- Prior experimental draft ScraperGo PR #78 and static contract PR #80 remain separate. Do not merge or release without owner approval.

# Active candidate — Issue #75 (2026-10-25)

- Base: current `main` containing merged PR #73; branch `fix/issue-75-nyt-preview-extraction`. Issue: https://github.com/richard-mauri/scrapergo/issues/75.
- Diagnosis: Safari captured HTML includes additional embedded `ParagraphBlock` data, but the extracted 372-word output mostly represents the rendered preview and includes publisher access UI. Hidden structured blocks must not be treated as public article text.
- Narrow implementation: `internal/reader/access_chrome.go` removes only identified access-check/login/subscribe UI paragraphs after Readability and serialized-player cleanup; retains publisher restriction notice and editorial content. Tests update the reader stage order and verify preservation of regular prose and images.
- Limitations: pattern matching is conservative and currently tied to specific access UI wording. No general embedded-JSON article recovery, no subscriber content reconstruction, no retrieval/NewsGo changes.
- CI and Mac/Safari acceptance NOT verified. Do not merge, tag, release, or bump semantic version without explicit owner authorization. Issue #74 performance follow-up and #69 image delivery are separate.

---

# Active candidate — Issue #72 (2026-10-25)

- Baseline: main; isolated branch: `fix/issue-72-shared-article-fetch` (draft PR #73 open: https://github.com/richard-mauri/scrapergo/pull/73).
- Issue: https://github.com/richard-mauri/scrapergo/issues/72
- Scope: share in-flight article HTML downloads across classification, thumbnail metadata and reader paths; keep successful HTML in a bounded 30-second in-memory cache (12 entries / 12 MiB), without caching errors or canceled downloads.
- Architectural boundary: the new coordinator lives in `internal/fetcher`; the underlying secure URL validation, redirect and transport logic stays in that package. No NewsGo API or UI changes.
- Tests added for simultaneous identical requests, cache hits, publisher denials/recovery, canceled waiters, canceled fetches, and oversized results.
- Validation: Go tests, exact-head CI and owner Mac/Safari acceptance not yet verified. No release or merge authorization; no semantic version change.
- Limitation: this reduces repeat requests but does not demonstrate why NYT temporarily returned 403.
- Other active issue #69 LA Times image delivery remains separate and unmerged. Do not combine branches.

---

# Current active work — Issue #69 (2026-10-25)

- Authoritative baseline: `main`; dedicated branch: `fix/issue-69-la-times-image-delivery` (not merged).
- Goal: avoid Brightspot 1x1 hotlink placeholders in the article reader without modifying NewsGo's default image API behavior.
- Safari evidence: original LA Times image URL directly opens a full photograph; embedded request redirects to `placeholder-1x1.png` and reports intrinsic 1x1.
- Changes: client retries missing/1x1 images via same-origin `/api/article-image`, with candidate index preserving distinct images; server validates explicit index and rejects 1x1 PNG/JPEG/GIF results.
- Tests: UI fallback assertions and HTTP invalid candidate index regression; Go CI and owner Mac/Safari acceptance pending.
- Regression URLs: https://www.latimes.com/sports/ucla/story/2026-10-25/ucla-perfect-start-glaring-flaw-bob-chesney and the two earlier UCLA links in Issue #69.
- Next: open draft PR, verify exact-head CI, owner Mac/Safari test, adjust if needed. Do not merge/tag/release or bump semantic version without owner authorization.
- Deferred: diagnose publishers that serve placeholders in other image formats; image API remains restricted to extractor-selected candidates.

---
# ScraperGo — Current work

## Current documentation-only branch (2026-10-24)

- Working branch: `docs/branch-lifecycle-20261024` → `main`. Scope: document safe post-merge branch deletion, prevent accumulated stale heads, preserve owner-controlled releases.
- Updated `DEVELOPMENT.md` while keeping `AGENTS.md` as its existing agent entry point.
- Existing historical branches are **not deleted**. ScraperGo currently has 66 remote branches, 63 matching heads of merged PRs; bulk cleanup requires a separately approved safety audit.
- No runtime code or user state changes. Check final-head CI before handoff; owner approval required to merge this documentation PR.
- Previous PR-cycle notes below are historical context and should be reconciled with live GitHub state, not treated as current open PRs.


**Snapshot:** October 23, 2026 · **Repository:** https://github.com/richard-mauri/scrapergo

## Active PR: #64 — documentation consolidation

- **PR:** https://github.com/richard-mauri/scrapergo/pull/64
- **Branch:** `docs/consolidate-reference-20261023` → `main`
- **Status last checked:** Open draft, not merged. Recheck latest GitHub PR state and branch head before changes.
- **Starting main baseline:** `eaa08b8dcefa1ce2b9232c3e38839905cc575065`, after merged PRs #53–#63; `appVersion` is `0.2.85`.
- **Last verified CI before recent documentation edits:** [run 40029538019](https://github.com/richard-mauri/scrapergo/actions/runs/40029538019) succeeded at commit `0440b797ec172d35347d9eae00c04936c2c1dd82`. This is **not** evidence for subsequent commits. New-head CI needs verification.
- **Objective:** A concise, accurate, human-readable documentation system reflecting merged implementation, with a durable single-file handoff for active PR cycles. This PR must remain documentation-only.

## Work completed in this PR

- Replaced outdated README development diary with product/usage/API documentation and an illustrative dark-theme SVG homepage preview (`docs/images/scrapergo-home-preview.svg`).
- Rewrote `ARCHITECTURE.md` for actual package ownership and the existing architecture boundary tests.
- Created `DEVELOPMENT.md` for branches, Issues, focused PRs, CI, Mac procedure and release authorization.
- Condensed and reformatted `AGENTS.md` as a readable agent entry point; removed duplicate process/architecture details.
- Updated `docs/repository-layout.md`, fixture test reference, and historical-status warnings on older architecture/design documents.
- Removed `HISTORY.md` at owner's request; existing Git history and `docs/archive/` preserve provenance.
- Introduced `CURRENT_WORK.md` to carry concrete active project context and out-of-scope findings between chats. Its maintenance instructions now belong in `AGENTS.md`.

## Decisions to preserve

- **One authoritative owner per topic:** README = product and usage; AGENTS = agent obligations and handoff maintenance; ARCHITECTURE = ownership/dependency contract; DEVELOPMENT = GitHub workflow and validation; CURRENT_WORK = actual active state and unresolved context.
- Keep `AGENTS.md` concise, scannable and human-readable.
- Do not restore `HISTORY.md` or rebuild old cumulative handoff diaries.
- Documentation-only changes need Markdown/link/path/content review and normal CI, **not** a local Mac build, application restart or Safari acceptance.
- No merge, tag, release or semantic-version change without owner approval. Preserve local runtime credentials, tunnel configuration and private build counter.
- The README illustration is an **SVG recreation**, not the exact screenshot.

## Outstanding tasks for PR #64

- [ ] Review these documentation changes for accuracy, clarity and unnecessary duplication.
- [ ] Check all internal Markdown links, commands and package/test references against the PR branch.
- [ ] Check the exact latest PR head and GitHub Actions result after the most recent edits.
- [ ] Obtain owner review and explicit merge authorization. **Do not merge automatically.**

**Immediate next action:** Verify the current draft PR #64 head and CI, inspect the final documentation-only diff, and present remaining issues to the owner for review.

## Deferred findings — outside PR #64

- [ ] Synthetic HTML extraction fixtures do not validate all live CNN/NPR or other publisher behavior. Investigate actual reported regressions separately using appropriately reviewed source evidence; record reproducible defects as GitHub Issues and fix in focused PRs.
- [ ] Any new unrelated reader, retrieval, image, Safari extension, thumbnail or Cloudflare defects discovered during acceptance belong in a separate Issue/PR, not this documentation cleanup.

## Local environment relevant to continuation

Mac checkout: `~/projects/scrapergo` (older Git supports `git checkout`, not `git switch`). Runtime local address: http://localhost:8083 when launched by the established ScraperGo launcher. Development/build/acceptance commands live in `DEVELOPMENT.md`.

**Continuity:** This document records what is known from the present PR cycle; current GitHub state must be reverified on resumption. As further PRs become active, record each branch, PR link, dependency, technical discoveries, next action and unresolved TODO here, replacing superseded status rather than accumulating history.
