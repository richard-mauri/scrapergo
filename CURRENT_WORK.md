# Current work — ScraperGo

**Active documentation review:** [Issue #138](https://github.com/richard-mauri/scrapergo/issues/138), branch `docs/138-session-markdown-refresh` targeting `main`, [draft PR #139](https://github.com/richard-mauri/scrapergo/pull/139). This branch changes only Markdown; exact-head CI and owner merge approval are pending.

GitHub `main` at `8ab1bb0` includes merged suite activity, verified Homepage guest/session reporting, immediate first-arrival activity, revoked-heartbeat handling and non-dismissible sign-in UI (PRs #133, #135, #137). Issues #134 and #136 are closed. No issues or PRs remain open and no implementation branch is active.

Homepage is the identity authority. ScraperGo verifies supplied bearers for optional activity; revoked identity returns 401, verification outage returns 503, and anonymous visits remain distinct. A verified NewsGo arrival promptly reaches the Homepage activity collector. The sign-in dialog cannot be dismissed with Close/Escape, while manual Homepage navigation remains available. Article extraction, SSRF protections, rendering, storage, listener and tunnel are unchanged. The broader API-authorization audit is separately tracked by [Homepage Issue #38](https://github.com/richard-mauri/homepage/issues/38).

**Next:** reconcile Markdown through a scoped docs-only draft PR from `main`; verify Markdown references and exact-head CI. Runtime acceptance on Safari/live publishers and merge, restart, release or tunnel actions remain owner-controlled. Recheck current GitHub and any owner-local divergence before resuming.
