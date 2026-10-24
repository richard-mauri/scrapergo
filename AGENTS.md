# ScraperGo — agent instructions

This is the **entry point for development agents**, not a duplicate of the architecture or contributor manuals. At the start of a session, read [CURRENT_WORK.md](CURRENT_WORK.md) for the active PR cycle, unresolved decisions and deferred TODOs. Then read:

- [ARCHITECTURE.md](ARCHITECTURE.md) — binding package ownership, dependencies, security boundaries and architecture tests.
- [DEVELOPMENT.md](DEVELOPMENT.md) — authoritative GitHub branch, Issues, PR, testing, Mac acceptance and release procedures.

## Mandatory workflow acknowledgment — before development

Before creating a branch, modifying source or documentation, opening an Issue or PR, or carrying out other development work, **read this file, CURRENT_WORK.md, ARCHITECTURE.md and DEVELOPMENT.md from the current GitHub working branch**. Check the actual GitHub branch and PR state; handoff notes may be stale.

**State the applicable workflow back to the owner before proceeding.** The acknowledgment must cover: (1) the authoritative merged baseline, active branch/PR and current work; (2) Issue, scoped-branch and draft-PR procedure; (3) code/test and architecture constraints; (4) CI checks and their limits; (5) the complete Mac checkout, test, build and launch commands with the real branch name for code/runtime work; and (6) owner-side Safari acceptance and explicit authorization before merge, tag or release.

Do not substitute a generic promise to follow instructions for this concrete summary. **At code/runtime handoff, repeat the entire copy-and-pasteable Mac command block from DEVELOPMENT.md with the actual branch name**, even if it was shown earlier. For documentation-only work, state the applicable documentation/CI checks instead. Clearly identify checks not performed and blockers; never imply CI validates Safari or a live publisher.

## Mandatory conduct

- Inspect the actual current code and relevant tests before proposing changes; do not follow obsolete migration notes as current instructions.
- Respect every architectural constraint in **ARCHITECTURE.md**. In particular, do not add migration-only compatibility wrappers or APIs merely to keep old callers or tests compiling.
- Preserve existing runtime behavior, regression coverage, fixtures and security protections unless a change is explicitly scoped and justified.
- Apply the bug, branch, PR and validation procedures in **DEVELOPMENT.md**; track unrelated discoveries as separate Issues.
- Do not silently widen a focused PR into unrelated refactoring or behavior changes.
- Report precisely which checks passed, failed or were not performed. Do not describe CI as Mac or Safari acceptance.
- **Do not merge, tag, release or change semantic version without the owner's explicit authorization.**
- Never expose or change private credentials, tunnel configuration or local build counters.

## Maintaining CURRENT_WORK.md

- Keep `CURRENT_WORK.md` as a **single portable snapshot of actual work**, not a template, policy manual or chronological diary. It must stand alone when uploaded into a fresh chat.
- Capture active PRs and branches (links, base/head, state, dependencies), current goal and technical findings, decisions and reasons, completed work, next action, blockers, verified test/CI evidence and unresolved out-of-scope TODOs.
- **Add** new findings and deferred tasks; **replace** outdated statuses, SHAs and next steps. Retain relevant decisions and TODOs across PR cycles until resolved, transferred to an Issue, or explicitly dismissed.
- Update the same file before a chat handoff and at meaningful PR transitions. Clearly mark unverified status; check GitHub again at the start of the next session. Never create a stack of dated handoffs.
- Keep reusable instructions here, architecture in `ARCHITECTURE.md`, and branching/testing policy in `DEVELOPMENT.md`. Do not duplicate them inside `CURRENT_WORK.md`.

## Handoff

**Before handing a PR to the owner for review, update the existing GitHub PR title and description to reflect the actual latest scope and implementation.** Replace obsolete diagnostic-only, proposed-work, test-pending, or other stale statements. Include the current head commit, relevant Issue links, changed files/behavior, exact CI result at that head, Mac/Safari acceptance evidence, outstanding checks, known limitations, and explicit draft/no-merge status. Verify the saved PR text on GitHub after the update; never hand off a stale PR description or claim unperformed checks passed.

Follow the change-type-specific validation in **DEVELOPMENT.md**. For code/runtime PRs, include its full Mac checkout, test, build and launch sequence with the **actual branch name** substituted. Documentation-only PRs require documentary review and existing CI, not a local application build or Safari acceptance.

Historical notes under `docs/` and in Git history are reference material, **not current instructions**.
