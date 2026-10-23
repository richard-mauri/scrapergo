# ScraperGo development and validation

Start with [AGENTS.md](AGENTS.md) and [ARCHITECTURE.md](ARCHITECTURE.md). GitHub [main](https://github.com/richard-mauri/scrapergo/tree/main) is authoritative. Keep branches short-lived and scoped (for example `docs/consolidate-reference`, `fix/issue-N-description`, `test/area-contract`). Open a **draft PR** against `main` with rationale, changed boundaries, issue references, regression risks, CI and Mac/Safari acceptance. The owner decides when to merge, tag or release. Do not bump `appVersion` during documentation or ordinary development PRs.

## Bugs and GitHub Issues

1. Reproduce the defect; open an Issue in the affected repository before broadening scope.
2. Record the Issue URL, expected versus observed behavior, exact reproduction, diagnostics, affected build, and relevant screenshots (redact tokens/private data).
3. Implement a focused PR referencing the Issue and add regression coverage. A defect discovered while working on unrelated changes normally becomes a separate Issue.
4. Verify GitHub Actions at the exact PR head, then perform owner-side Mac/Safari acceptance.
5. Obtain owner approval before merge. Use `Fixes #N` only if the PR actually resolves that Issue.

A cross-repository example is [NewsGo Issue #8](https://github.com/richard-mauri/newsgo/issues/8): track the bug where it belongs, rather than slipping unrelated NewsGo work into ScraperGo.

## Automated and local checks

GitHub Actions workflow [Go CI](.github/workflows/go-ci.yml) runs on PRs targeting `main`: a non-blocking gofmt debt report, `go test -count=1 ./...` and `go build ./...` (Go version from `go.mod`). Inspect the run and outcome on **this PR's latest commit**. Confirm the intended diff is documentation-only when applicable.

For **every** PR, supply the following complete copy-and-pasteable block to the owner with its *real branch name*, not the placeholder:

```bash
cd ~/projects/scrapergo
git fetch origin
git checkout <exact-pr-branch>
git pull --ff-only origin <exact-pr-branch>
go test -count=1 ./...
./devbuild.sh
./scrapergo.sh
```

Use `git checkout` for the owner's older Git. Do not add a separate `go build` to this Mac sequence: `devbuild.sh` already runs tests and builds Intel `dist/scrapergo-darwin-amd64` and Apple Silicon `dist/scrapergo-darwin-arm64` binaries and SHA256SUMS. Its sequential build counter is stored privately under the local Git directory (`git rev-parse --git-path scrapergo-build-counter`) and is independent of semantic version and PR number. Never reset or commit it. The established `scrapergo.sh` launcher selects the Intel binary, runs the server at 8083 and starts **only** ScraperGo's Cloudflare tunnel; it depends on local private token and binary paths. Do not change tunnel credentials, machine-specific configuration, launch process isolation or other applications. `./devbuild.sh --release vX.Y.Z` publishes a GitHub release and is **not** a normal development step; explicit owner authorization is required.

## Coverage and Safari acceptance

Check source-backed media identity, editorial lead versus author portrait, caption/order, contributor bylines, article/partial-page classification, links, access denial, error responses and thumbnails. Preserve test fixtures (including `testdata/` and package fixtures); see [docs/extraction-fixtures.md](docs/extraction-fixtures.md). The frozen HTML examples are synthetic, not captured publisher behavior. Do not replace behavioral contracts just to make tests pass. Use separately reviewed and permission-appropriate real publisher captures where needed; avoid including account data or copyrighted full-page material unnecessarily.

After local build/launch, inspect **http://localhost:8083** in Safari, including the reading view, extension link handoff where applicable, image display and print/PDF behavior. Live publisher, Cloudflare and Safari checks are owner-side acceptance, not evidence supplied by Linux CI. Preserve the current browser and [Safari extension setup](safari-extension/README.md). Report checks as passed, failed or not performed; never imply a remote CI run substitutes for browser acceptance.

For historical architectural changes see [HISTORY.md](HISTORY.md); historical plans in `docs/` are not current development instructions.
