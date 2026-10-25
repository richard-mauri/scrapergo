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
