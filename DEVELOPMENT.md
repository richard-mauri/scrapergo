# ScraperGo — app-specific development and validation

Read [AGENTS.md](AGENTS.md) and the full [generated suite policy](docs/suite-development-policy.md). Homepage owns the shared workflow, six-item handoff, FINISH, and branch cleanup. The following are ScraperGo-specific extraction gates, tests, builds and Safari acceptance.

## General-purpose fix acceptance gate

A successful screenshot from the triggering publisher is **not** evidence of a general fix. Before editing production behavior, capture or inspect the source/Readability representations using [the offline probe and optional secure capture](cmd/readabilityprobe/README.md), document where information disappears, and state a publisher-independent invariant. Prefer semantic DOM evidence over URL substrings, publisher branding, relative-time regexes and text adjacency. If a site-specific exception is genuinely needed, obtain explicit owner approval and document why the general solution is insufficient.

Before handing off, require regression cases for the original failure; an unrelated ordinary article; a second distinct article/update structure; and a negative or ambiguous case where the extractor must **not** fabricate headings, timestamps or boundaries. Verify canonical article blocks and rendered/exported output when applicable, including links, paragraphs, images and source order. Distinguish synthetic-fixture evidence, actual HTML capture, server extraction, and Safari acceptance; none substitutes for the others. If real-site capture is unavailable, state the remaining uncertainty rather than claiming success.

Stay on the existing Issue branch and draft PR for revisions; do not spin up additional branches or PRs to try another heuristic. Prefer completing one general-purpose solution end to end, avoiding parallel branch proliferation. Changes to the architecture or scope require owner approval when they affect unrelated behavior.

## Automated and local checks

GitHub Actions workflow [Go CI](.github/workflows/go-ci.yml) runs on PRs targeting `main`: a non-blocking gofmt debt report, `go test -count=1 ./...` and `go build ./...` (Go version from `go.mod`). Inspect the run and outcome on **this PR's latest commit**. Confirm the intended diff is documentation-only when applicable.

For **code or runtime PRs**, supply the following complete copy-and-pasteable block to the owner with its *real branch name*, not the placeholder. **Documentation-only PRs** instead require rendered Markdown, link, path and command verification; normal PR CI can run, but local Mac build/restart/Safari acceptance is unnecessary:

```bash
cd ~/projects/scrapergo
git status --short --branch
# Stop if uncommitted changes are shown.
git fetch origin
git checkout ACTUAL-PR-BRANCH
git pull --ff-only origin ACTUAL-PR-BRANCH
go test -count=1 ./...
./devbuild.sh
./dist/scrapergo-darwin-arm64 --version
./scrapergo.sh
```

The normal `./devbuild.sh` runs a short `bash ./benchmark.sh --smoke` check after ordinary Go tests; benchmark execution errors fail the build, but timing variation is not a failure. Full timing measurements remain opt-in via `./benchmark.sh` and are not part of plain `go test ./...` or routine GitHub Actions.

Use `git checkout` for the owner's older Git. Do not add a separate `go build` to this Mac sequence: `devbuild.sh` already runs tests and builds Intel `dist/scrapergo-darwin-amd64` and Apple Silicon `dist/scrapergo-darwin-arm64` binaries and SHA256SUMS. Use the **ARM64** executable for version checking and launching on Richard's Apple Silicon Mac; AMD64 is compatibility output for Intel Macs. Its sequential build counter is stored privately under the local Git directory (`git rev-parse --git-path scrapergo-build-counter`) and is independent of semantic version and PR number. Never reset or commit it. The established `scrapergo.sh` launcher selects the Apple Silicon `dist/scrapergo-darwin-arm64` binary, loads the private `.activity.env`, runs the server at 8083 and starts **only** ScraperGo's Cloudflare tunnel; it depends on local private token and binary paths. Do not change tunnel credentials, machine-specific configuration, launch process isolation or other applications. `./devbuild.sh --release vX.Y.Z` publishes a GitHub release and is **not** a normal development step; explicit owner authorization is required.

## Coverage and Safari acceptance

Check source-backed media identity, editorial lead versus author portrait, caption/order, contributor bylines, article/partial-page classification, links, access denial, error responses and thumbnails. Preserve test fixtures (including `testdata/` and package fixtures); see [docs/extraction-fixtures.md](docs/extraction-fixtures.md). The frozen HTML examples are synthetic, not captured publisher behavior. Do not replace behavioral contracts just to make tests pass. Use separately reviewed and permission-appropriate real publisher captures where needed; avoid including account data or copyrighted full-page material unnecessarily.

After local build/launch, inspect **http://localhost:8083** in Safari, including the reading view, extension link handoff where applicable, image display and print/PDF behavior. Live publisher, Cloudflare and Safari checks are owner-side acceptance, not evidence supplied by Linux CI. Preserve the current browser and [Safari extension setup](safari-extension/README.md). Report checks as passed, failed or not performed; never imply a remote CI run substitutes for browser acceptance.

Historical plans in `docs/` and previous Git revisions are not current development instructions.
