# Current work

**Authoritative merged baseline:** GitHub `main` at `94365386e010e1a37fdaf9ded7b8951941a8fcf2` (November 6, 2026), after owner-authorized PR #115 squash merge; issue #114 closed. Earlier merged PR #109 established source-backed Readability heading/time restoration and was owner-accepted. The owner's primary cleanup goal is identifying obsolete, brittle recovery mechanisms after generalized HTML processing, **not** performance optimization.

**Sole active development task:** [Issue #117](https://github.com/richard-mauri/scrapergo/issues/117) persistent extraction recovery tracing. Branch `diag/117-persistent-recovery-trace`, [draft PR #118](https://github.com/richard-mauri/scrapergo/pull/118) against `main`. This is instrumentation only, not removal of recovery functions. [Issue #116](https://github.com/richard-mauri/scrapergo/issues/116) shared-DOM parsing/performance optimization is deferred, with the owner's measured Intel baseline recorded in the issue. Do not create a parallel branch/PR for #116.

**Scope on #117:** `internal/extraction/recovery_trace.go` and `ExtractWithReadabilityTraced` observe existing recovery stages, including in-place link modifications and legacy/partial fallback selection. Plain `ExtractWithReadability` retains its signature and executes the same operations without tracing. `internal/diagnostics/recovery_store.go` appends privacy-minimized observations to `diagnostics/recovery.jsonl` under the current working directory when enabled, using a persistent keyed URL identifier from a private `.recovery-key`. `internal/bootstrap/service.go` enables the recorder only if `SCRAPERGO_RECOVERY_TRACE=1`. `scrapergo.sh` opts in explicitly and already changes to `/Users/richardmauri/projects/scrapergo` so storage lands at the owner-selected absolute path. `.gitignore` excludes `/diagnostics/`. `cmd/recoveryreport` prints cross-session counts without URLs or article text. Tests cover in-place mutations, extraction output equality, persistence and privacy.

**Limitations:** Records only extraction requests reaching ScraperGo, not browser navigation that does not call the service. A stage's `modified` status means the immediate step changed canonical blocks and does not prove the change survives later stages. No article text/full URLs are retained; keyed article IDs require the persistent `.recovery-key`. Existing extraction security/access checks and fallback behavior must remain unchanged. No real publisher or Safari comparison has been performed yet.

**Validation:** GitHub Actions run [42282807253](https://github.com/richard-mauri/scrapergo/actions/runs/42282807253) passed tests/build at PR head `e0c1ef73abe86f40a14e976431d19479519cc930`. Exact-head CI must be reconfirmed after any follow-up commit. Owner Mac `./devbuild.sh`, `./scrapergo.sh`, Safari testing via Homepage/NewsGo/direct ScraperGo, repeated requests, restart continuity and `go run ./cmd/recoveryreport`. Ensure user sees logs under `diagnostics/`. PR remains draft; no merge, release, tag, version change or branch deletion without explicit owner approval.

**Mac command sequence once PR is ready:**

```sh
cd ~/projects/scrapergo
git fetch origin
git checkout diag/117-persistent-recovery-trace
git pull --ff-only origin diag/117-persistent-recovery-trace
go test -count=1 ./...
./devbuild.sh
./scrapergo.sh
```

Then visit articles in Safari via Homepage and NewsGo and directly in ScraperGo; verify article UI/links/figures/print and that repeated requests increment `diagnostics/recovery.jsonl`, including after restarting the app. Run `go run ./cmd/recoveryreport` to see aggregate counts.

**Deferred:** Behavioral removal candidates include `RestoreHeadingAnchoredProse`, `RecoverEditorialContinuity`, and `RecoverInstitutionalTail`, but rarely invoked code is not automatically dead. Future targeted ablation/equivalence tests and publisher-diverse evidence are required before any deletion. #116 performance optimization remains deferred.
