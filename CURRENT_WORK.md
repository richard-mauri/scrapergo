# Current work — ScraperGo

## Issue #136 — Non-dismissible suite sign-in prompt

- **Baseline:** GitHub `main`, including merged ScraperGo PRs #133 and #135 (Homepage guest handoffs and first verified heartbeat).
- **Branch:** `fix/136-block-signin-dismiss` from `main`; focused draft PR pending.
- **Owner-reproduced UX:** Admin revokes a temporary guest; ScraperGo correctly reports sign-out but its modal's **Close** button let an unidentified visitor continue reading articles.
- **Change:** Remove **Close** from the sign-in-required dialog; prevent native Escape/cancel dismissal. Preserve explicit **Go to Homepage** link and successful Homepage verification's programmatic dismissal, without automatic navigation.
- **Boundaries:** No changes to ScraperGo extraction, request handlers, public API authorization, optional telemetry cadence, private tokens, launchers or publisher safety. Homepage Issue #38 separately tracks a deferred API authorization audit.
- **Regression:** Source-contract test in `internal/activity/activity_test.go`; exact-head Go CI plus owner Apple Silicon Safari acceptance pending. CI is not Safari acceptance.
- **Owner acceptance:** Revoked or missing identity shows modal with only **Go to Homepage**; Escape cannot dismiss; verified identity restores reader UI. No merge, tag, release or service restart without explicit owner approval.

Mac commands:

```bash
cd ~/projects/scrapergo
git fetch origin
git checkout fix/136-block-signin-dismiss
git pull --ff-only origin fix/136-block-signin-dismiss
go test -count=1 ./...
./devbuild.sh
./dist/scrapergo-darwin-arm64 --version
./scrapergo.sh
```
