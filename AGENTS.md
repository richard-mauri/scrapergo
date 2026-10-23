# ScraperGo — agent instructions

This is the **entry point for development agents**, not a duplicate of the architecture or contributor manuals. At the start of a session, read [CURRENT_WORK.md](CURRENT_WORK.md) for the active PR cycle, unresolved decisions and deferred TODOs. Then read:

- [ARCHITECTURE.md](ARCHITECTURE.md) — binding package ownership, dependencies, security boundaries and architecture tests.
- [DEVELOPMENT.md](DEVELOPMENT.md) — authoritative GitHub branch, Issues, PR, testing, Mac acceptance and release procedures.

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

Follow the change-type-specific validation in **DEVELOPMENT.md**. For code/runtime PRs, include its full Mac checkout, test, build and launch sequence with the **actual branch name** substituted. Documentation-only PRs require documentary review and existing CI, not a local application build or Safari acceptance.

Historical notes under `docs/` and in Git history are reference material, **not current instructions**.
