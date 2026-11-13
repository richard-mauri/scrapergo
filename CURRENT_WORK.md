# Current work — ScraperGo

## Issues #136 and Homepage #39 — Guest sign-in modal and immediate arrival activity

- **Baseline:** GitHub `main`, including merged ScraperGo PRs #133 and #135 (Homepage guest handoffs, verified first heartbeat).
- **Branch/PR:** `fix/136-block-signin-dismiss`, draft PR #137, targeting `main`. The owner authorized completing the full Homepage→apps, NewsGo→ScraperGo, and apps→Homepage transition acceptance before merging; companion Homepage draft PR #40.
- **Sign-in UX:** Remove **Close** from the sign-in-required dialog and prevent native Escape/cancel dismissal. **Go to Homepage** remains a user-controlled link, with no automatic redirects. Verified Homepage identity closes the modal.
- **Arrival telemetry:** The already authenticated `POST /api/activity/ping` invokes the existing server-to-server report promptly on an accepted verified ping, including after NewsGo opens ScraperGo directly. It does not introduce cross-origin JavaScript, new identities, new secrets, token renewal or a shorter browser heartbeat. The normal 60-second ticker restarts after an immediate event to avoid duplicate scheduled reports. The Homepage collector still owns per-user activity diagnostics.
- **Boundaries:** No alteration to article APIs, extraction/SSRF, persistence, versioning, private credentials, launcher or tunnel. Broader API guarding remains deferred to Homepage Issue #38.
- **Checks:** Regression tests for non-dismissible dialog and immediate verified report. Exact-head CI and owner Mac/Safari acceptance required; CI does not prove live Safari behavior.
- **Owner test:** From a verified Homepage guest, visit NewsGo then click an article that opens ScraperGo. In admin Users & Access, expect a ScraperGo last-seen/current observation promptly, without waiting a full minute. Repeat from an existing ScraperGo tab; test invalid/revoked guest popup and manual Back to Homepage. Do not merge without explicit owner approval.

Mac commands:

```bash
cd ~/projects/scrapergo
git status -sb
git fetch origin
git checkout fix/136-block-signin-dismiss
git pull --ff-only origin fix/136-block-signin-dismiss
go test -count=1 ./...
./devbuild.sh
./dist/scrapergo-darwin-arm64 --version
./scrapergo.sh
```
