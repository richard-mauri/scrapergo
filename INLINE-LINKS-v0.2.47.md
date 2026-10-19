# ScraperGo v0.2.47 — Inline links and lists

One authoritative version constant (`appVersion`) is now `0.2.47`.

The general HTML block extractor retains inline anchor URLs, bold and italic text parts, and separate HTML `<li>` elements. Relative links are resolved against the article address. Only HTTP and HTTPS link targets are made clickable. Short paragraphs containing links are no longer discarded.

Existing browser rendering of `parts` is reused; its Markdown export now includes inline Markdown links. No publisher-specific selectors were added. The image recovery/deduplication code remains unchanged.

The Sanders press release is a real-world manual verification case: confirm its bill text, summary and economic analysis PDF links are clickable and the funding items remain separate.

## Testing limitation

The available ScraperGo source snapshot omits pre-existing helpers used by the HTTP read handler. The patch's tests and inherited tests passed in an isolated harness with temporary, unshipped test stubs; full integration must be checked in the complete local repository with `go test ./...` and `./devbuild.sh`.
