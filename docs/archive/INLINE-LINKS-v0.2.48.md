# ScraperGo v0.2.48 — Inline-link recovery

Fixes a link-loss path where Readability's plain-text extraction or structured-document recovery returned selected article paragraphs without their original hyperlink metadata.

`recoverMissingArticleLinks` walks original downloaded HTML without executing JavaScript, matches paragraph/list text to already-selected article blocks, and attaches safe HTTP(S) links only. It excludes navigation, header, footer, and other noisy content. Existing rich-link parts remain unchanged.

The browser opens PDF and other document links directly, while HTML article hyperlinks continue to use ScraperGo. The single `appVersion` constant is 0.2.48. Image extraction and deduplication are unchanged.

The dedicated tests were run using an isolated harness because the complete Go project and its Readability dependency are not present in this environment. Please run `go test ./...` and `./devbuild.sh` on the complete Mac checkout. Verify the three PDF anchors in `/api/read` as well as Safari.
