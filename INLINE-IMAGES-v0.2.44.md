ScraperGo v0.2.44 diagnostic patch

Recovers image blocks from editorial image wrappers which may fall outside the text-selected article subtree. Rejects noisy / navigation content; preserves current article text. Includes a synthetic FOX-shaped wrapper regression test.

This will NOT recover photographs inserted exclusively by JavaScript into a DOM after initial page load. If the original HTTP response does not include the photograph URLs, an independent rendered-page retrieval strategy is necessary. Run go test ./... on the complete local project before deploying. The source snapshot used for this patch lacks existing helpers and could not be fully compiled in this environment.
