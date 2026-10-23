# ScraperGo — contributor and agent instructions

These are **mandatory operating rules** for human and AI contributors. For details, see [ARCHITECTURE.md](ARCHITECTURE.md) and [DEVELOPMENT.md](DEVELOPMENT.md).

## Before starting

- Read all three documents above, then inspect current code.
- Treat GitHub `main` as authoritative for merged production code.
- Use a short-lived branch and a focused **draft PR**. Historical notes in `docs/` and Git history are not current instructions.

## Architecture and behavior

- Keep `main.go` as the executable entry point. Domain implementations and unit tests belong in their owning `internal/` packages; integration tests belong at the relevant service boundary.
- Maintain intentional, acyclic dependencies. `internal/reader` must not import HTTP, UI, or fetcher implementations.
- Use the established validated fetcher/Google News retrieval boundary. Do not add independent network retrieval in processing stages.
- **Do not introduce compatibility forwarding wrappers**, aliases, obsolete facades, or test-only APIs to keep old callers compiling. Fix callers and move tests to their real owners.
- Preserve existing behavior, security controls, fixture contracts, and regression coverage. Explain intentional fixture changes; passing tests alone does not prove architecture or live-publisher correctness.

## Issues and pull requests

- Open a GitHub Issue for a discovered bug or deferred enhancement; ordinarily keep unrelated discoveries out of the current PR.
- Capture reproduction steps, observed versus expected results, diagnostic evidence, screenshots, and the Issue URL.
- Reference the Issue from a focused PR; use `Fixes #N` only when the PR genuinely resolves it.
- Avoid mixing documentation, architecture, and behavioral changes without a clear reason.
- Verify GitHub Actions on the **exact PR head**, and accurately report passed, failed, pending, and owner-only checks.

## Validation: match the change

- **Code or runtime changes:** run `go test -count=1 ./...` and `go build ./...` (also run by the PR CI workflow). Supply the complete Mac command sequence below for owner acceptance; request Safari testing when runtime/browser behavior may be affected.
- **Documentation-only changes:** verify the Markdown rendering, links, commands, paths, and factual accuracy. Let existing CI run; **do not require** a local Mac build, restart, or Safari acceptance just for Markdown.
- The owner's older Git requires `git checkout`, not `git switch`.
- `./devbuild.sh` tests and builds both macOS architectures with a private sequential build counter; a separate Mac `go build` command is redundant. `./scrapergo.sh` launches the Intel binary and ScraperGo-specific Cloudflare tunnel.

## Authorization and safety

- **Never merge, tag, release, or change semantic application version without owner authorization.** Documentation-only changes must leave `appVersion` at `0.2.85`.
- Never disclose, reset, commit, or modify private credentials, tunnel configuration, or the local build counter.
- Do not run `./devbuild.sh --release` unless expressly authorized.

## Mac acceptance for code/runtime PRs

Provide this entire copy-and-pasteable block with the **actual branch substituted**. It is **not required for documentation-only PRs**.

```bash
cd ~/projects/scrapergo
git fetch origin
git checkout <exact-pr-branch>
git pull --ff-only origin <exact-pr-branch>
go test -count=1 ./...
./devbuild.sh
./scrapergo.sh
```

For affected browser workflows, the owner validates Safari at http://localhost:8083.
