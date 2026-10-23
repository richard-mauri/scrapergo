# ScraperGo — CURRENT_WORK handoff

> **Live PR-cycle snapshot — last updated 2026-10-23.**
> This is the **single file to upload to a new ChatGPT conversation** to resume this repository's active work. It describes active work and decisions, not an enduring changelog. **Verify live GitHub state first**: branches, PR state, CI and heads can change after this file is written.

## Resume instructions (read this first)

1. Use this document to reconstruct the current task, constraints, unresolved decisions and deferred work. Do **not** assume that open/merged/CI statuses are still current.
2. Fetch GitHub `main`, inspect the active PR(s), and compare commit SHAs before editing. GitHub `main` is the authority for merged code; this file is the authority only for the *recorded work intent and handoff*, not for code facts.
3. Read [AGENTS.md](AGENTS.md), [ARCHITECTURE.md](ARCHITECTURE.md), and [DEVELOPMENT.md](DEVELOPMENT.md) from the relevant branch for governing rules. In a new chat without repository access, this file should still tell you the immediate task and what to verify.
4. Continue from the latest verified PR state, not from an old chat's candidate or provisional commit. Do not create duplicate PRs or treat draft as merge authorization.
5. Update **this file** at significant handoffs, whenever PR scope changes, and before a chat transition. Replace stale statements rather than appending diary entries. Keep deferred TODOs until completed or explicitly dismissed.

## Project and standing constraints

- Repository: https://github.com/richard-mauri/scrapergo
- Local Mac checkout: `~/projects/scrapergo`; older Git requires `git checkout` rather than `git switch`.
- App version: `0.2.85`. No semantic bump, merge, release, tags, or release build without owner authorization.
- Architecture baseline: merged PRs #53–#63 and GitHub `main` commit `eaa08b8dcefa1ce2b9232c3e38839905cc575065` **at the start of this documentation PR**. Verify current `main` rather than treating that SHA as permanently current.
- Runtime launcher `./scrapergo.sh` uses the Intel Mac binary, local port 8083, and the dedicated Cloudflare tunnel; `./devbuild.sh` manages a private build counter. Never alter or expose secrets/counters.

## Active PR cycle

**PR #64: Documentation consolidation (DRAFT, not merged at this snapshot)**

- PR: https://github.com/richard-mauri/scrapergo/pull/64
- Branch: `docs/consolidate-reference-20261023` → `main`
- Last verified head *before this handoff edit*: `0440b797ec172d35347d9eae00c04936c2c1dd82`; **fetch the new head after editing this document**.
- Last checked CI for that preceding head: [run 40029538019](https://github.com/richard-mauri/scrapergo/actions/runs/40029538019), completed successfully. A new commit invalidates this as proof for the new head; check its CI separately.
- Goal: current, concise, human-readable docs grounded in merged source code, not migration-era instructions. Keep PR documentation-only.
- Files/intent: README (user-facing feature/installation/API docs and illustrative SVG hero); AGENTS (short agent rules and links); ARCHITECTURE (actual package boundaries); DEVELOPMENT (workflow and verification); docs/repository-layout (current files/tests); docs/extraction-fixtures (correct file path); historical design docs marked historical.
- Owner decisions already made: **do not retain HISTORY.md**; Git history and existing `docs/archive/` suffice. Make AGENTS visually scannable. Give each policy exactly one authoritative file (GitHub branch authority and CI/Mac workflow in DEVELOPMENT; architecture in ARCHITECTURE; product usage in README). Documentation-only changes do **not** require a local Mac build, restart or Safari acceptance. Do not accidentally revert these choices.
- README hero is a local illustrative SVG (`docs/images/scrapergo-home-preview.svg`), **not an exact screenshot**. User originally requested visual interest based on a screenshot; only change this asset on explicit request.
- Next step: review the newly added handoff file and its AGENTS/DEVELOPMENT links, verify links/paths and documentation-only diff, inspect CI on the *latest* PR head, send owner draft PR for review; **do not merge**.

## Current PR checklist

- [x] Consolidate README, AGENTS, ARCHITECTURE and DEVELOPMENT.
- [x] Correct outdated repository layout and mark older migration/design docs as historical.
- [x] Remove HISTORY.md and its live cross-references.
- [x] Add illustrative README image.
- [x] Make AGENTS scannable; separate architecture and development policy ownership.
- [x] Clarify docs-only validation versus code/runtime Mac/Safari acceptance.
- [ ] Review this CURRENT_WORK handoff for scope accuracy and appropriate upkeep rules.
- [ ] Validate internal Markdown links, documentation command examples, and no stale operational contradictions.
- [ ] Confirm GitHub Actions status for **latest PR head** after this update.
- [ ] Owner reviews PR #64 and decides whether/when to merge.

## Deferred/out-of-scope TODOs — do not lose these

Keep these visible but **do not implement them in PR #64** without express scope change:

- [ ] Live publisher acceptance gaps: synthetic frozen HTML fixtures cannot guarantee current CNN/NPR or other publisher image/byline behavior; create separate issues/PRs for proven defects with suitable captures.
- [ ] Any newly discovered runtime, retrieval, Safari, thumbnail, or Cloudflare defects: log in the *affected repository's* GitHub Issues with reproduction and diagnostics; fix under a focused PR.
- [ ] Review and update the handoff when PR #64 merges; remove closed PR-specific checkboxes, record the next active branches/PRs and carry forward unresolved TODOs. Never let this file become a second HISTORY.md.
- [ ] If several PRs are active simultaneously, add a separate small subsection for **each** (branch, base, PR link, exact head, state, dependency/order and next action). Do not infer that #64 is the only PR without checking GitHub.

## Owner approval and verification boundaries

The owner approves merges and releases. For this docs-only PR, check Markdown rendering, links, commands, paths and CI; **do not request a local build or Safari acceptance solely for documentation**. For future code/runtime PRs, DEVELOPMENT.md specifies the complete Mac commands and when Safari testing is needed. The owner uses the local browser at http://localhost:8083.

## Updating this handoff for the next chat

Before handing off, revise the **existing** CURRENT_WORK.md, rather than creating timestamped copies. Update date, all active PRs/branches and exact verified heads, CI outcome/links, what changed in the current cycle, remaining owner approvals, next specific action, blockers and out-of-scope TODOs. Cite issues/PRs with URLs. State unverified facts as unverified. A new chat should be able to resume from this one document without reconstructing old chat messages.
