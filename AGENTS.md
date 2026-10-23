# ScraperGo contributor and agent instructions

Read this file, [ARCHITECTURE.md](ARCHITECTURE.md), and [DEVELOPMENT.md](DEVELOPMENT.md) before changing anything.
GitHub `main` is authoritative for merged production code; fetch and inspect it before starting.
Use short-lived topic branches and focused draft pull requests; never work from an obsolete handoff.
Review code and package ownership rather than copying historical migration instructions.
Keep `main.go` an entry point only; domain logic and its unit tests belong in `internal/` owning packages.
Place integration tests at the actual service or application boundary, not automatically in root.
Maintain intentional, acyclic dependency direction; reader must not import HTTP, UI or fetcher.
Keep all remote retrieval within the existing validated fetcher/Google News resolution boundary.
Never add compatibility forwarding wrappers, aliases, obsolete facades or test-only APIs to preserve old callers.
Move actual callers and tests to the owning implementation instead.
Preserve existing production behavior, security controls, fixtures and regression coverage.
Treat existing fixture expectations as behavior contracts; explain any intentional update.
Create a GitHub Issue for discovered bugs or deferred enhancements; do not expand an unrelated PR silently.
Record reproducible symptoms, expected behavior, evidence, screenshots and a link to the Issue.
Reference the Issue in the fixing PR; use `Fixes #N` only when genuinely resolved.
Do not mix architecture, bug fixes, behavior changes and docs cleanup without justification.
Run `go test -count=1 ./...` and `go build ./...`; examine GitHub Actions on the exact PR head.
For every PR, give the owner the **complete** Mac checkout, test, build, launch command block below.
Replace `<exact-pr-branch>` with the real PR branch before sending; do not leave placeholders.
Use `git checkout`, not `git switch`, for the owner's older Git installation.
`./devbuild.sh` already tests and builds both Mac architectures with a local private build counter.
Do not invoke `./devbuild.sh --release` without explicit release authorization.
`./scrapergo.sh` launches the local Intel binary and the ScraperGo-specific Cloudflare tunnel.
Do not disclose, change, commit, reset or replace private tokens, tunnel configuration or build-counter files.
Safari acceptance is at http://localhost:8083 and remains an owner-side step.
Never merge a PR, create tags/releases, or change semantic version without owner authorization.
Documentation-only changes must not bump the current application version 0.2.85.
Report precisely what was tested and what requires owner verification.
Do not claim all architecture or publisher compatibility concerns have been solved.
Current architecture: [ARCHITECTURE.md](ARCHITECTURE.md).
Contributor procedure: [DEVELOPMENT.md](DEVELOPMENT.md).
Historical rationale: [HISTORY.md](HISTORY.md); prior notes are not current instructions.

## Required owner-side PR validation

```bash
cd ~/projects/scrapergo
git fetch origin
git checkout <exact-pr-branch>
git pull --ff-only origin <exact-pr-branch>
go test -count=1 ./...
./devbuild.sh
./scrapergo.sh
```
