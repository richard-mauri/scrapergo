# Current work — ScraperGo

**Active:** Homepage [Issue #45](https://github.com/richard-mauri/homepage/issues/45) consumer policy adoption on `docs/45-suite-policy-adoption` from GitHub `main`; [draft PR #140](https://github.com/richard-mauri/scrapergo/pull/140) open against `main`. Homepage `main` is the sole policy source; generated copies and sync checks removed. No consumer merge authorized.

**Merged context:** ScraperGo verifies Homepage guest/named identity for optional activity; revoked identities return 401 and verification outages 503. The sign-in dialog is non-dismissible and NewsGo arrival activity can be recorded promptly. Article extraction, publisher-specific invariants, SSRF controls, fixtures, Safari extension, renderer, listener and tunnel are unchanged. Protected API audit remains [Homepage Issue #38](https://github.com/richard-mauri/homepage/issues/38).

**Next:** review app-specific extraction guidance, Homepage policy link, Markdown and CI on the exact draft PR head, together with the other consumer PRs. Documentation-only work does not require live publisher/Safari retest or service restart. Preserve owner-local work and all private credentials/counters; no merge/release/tag/deploy.
