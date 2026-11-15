# ScraperGo — agent entry point

**Mandatory:** Read the full [generated local suite policy](docs/suite-development-policy.md) before any changes. This is a generated, pinned offline copy of Homepage's [authoritative shared policy](https://github.com/richard-mauri/homepage/blob/main/SUITE_DEVELOPMENT_POLICY.md). Verify locally with `python3 scripts/suite_policy.py --offline`; verify freshness online with `python3 scripts/suite_policy.py`. Do not hand-edit the copy; use `--sync` only in a reviewable PR. Missing/stale policy blocks shared-process edits.

Also read [README.md](README.md), [ARCHITECTURE.md](ARCHITECTURE.md), [DEVELOPMENT.md](DEVELOPMENT.md), and [CURRENT_WORK.md](CURRENT_WORK.md), and inspect GitHub and owner-local work. The suite policy owns workflow, six-part handoff, FINISH and branch cleanup. DEVELOPMENT owns extraction-specific acceptance and Mac commands.

ScraperGo's engineering contract requires publisher-independent evidence, source/Readability inspection and meaningful positive/negative regressions before changing extraction heuristics. Preserve SSRF controls, article classification, media identity, link/order/print rendering, Safari extension setup and test fixtures. Do not loosen extraction to hide upstream errors, publish copyrighted captures indiscriminately, change another service or touch private tokens/tunnel/build counters.
