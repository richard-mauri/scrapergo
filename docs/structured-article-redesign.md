> **Historical design record — not current instructions.** This document describes earlier migration/design states and may reference superseded branches, root package owners, aliases, and PR status. Current authority: [ARCHITECTURE.md](../ARCHITECTURE.md), [DEVELOPMENT.md](../DEVELOPMENT.md), and [HISTORY.md](../HISTORY.md).

# Structured Article Extraction Redesign

Status: staged implementation v0.2.73; **not yet a Mac release candidate**. Structural media DOM sanitization and opening-title / identical-image reconciliation have been implemented and tested, and the accepted NewsGo-inspired palette and centered spacing have been brought over without experimental image heuristics. Full representative publisher fixtures, robust source image-role classification, and browser/PDF regression coverage remain outstanding.

## Branch policy
- `main` remains production/stable; do not merge until owner explicitly accepts a tested candidate.
- `archive/scrapergo-v0272-experiments` retains all v0.2.66–v0.2.72 palette, media suppression, embed and title heuristics unchanged for comparison and selective salvage.
- `release/newsgo-palette-v0266` and its draft PR #3 are superseded; no squash-merge of that combined history.
- `rewrite/structured-articles` starts at `main`. Bring over accepted color/spacing separately from functional extraction changes, and only after baseline tests.
- `dev` is retained as-is and will be reconciled after the rewrite is validated; do not force-push or delete it.

## Root causes established
1. The reader renders a separate `imageUrl` hero and `blocks[].img` photos independently; arbitrary URL or image-count guesses hide legitimate leads or duplicate photographs.
2. Readability can emit escaped embed markup fragmented among text blocks. Post-hoc removal of neighboring blocks is brittle and may discard prose.
3. `Article.Title` and opening content headings can both represent the same editorial title.
4. UI-string presence tests and passing `go build` cannot establish actual rendered behavior.

## Target contract
Introduce a normalized article document consumed by a single renderer. It must retain stable image identities and explicit roles: lead, editorial figure, author portrait, decorative image. A lead image must be rendered exactly once with its caption/credit, not inferred from URL spellings, dimensions after layout, or the mere existence of inline photographs. Metadata and content title are reconciled at the document level.

Remove interactive media containers and descendants at the DOM/structured extraction boundary before text serialization, without using proximity-based block deletion. Recognize native iframe, audio, video, and embed nodes, plus *escaped markup that genuinely encodes a player*, but preserve surrounding article paragraphs and captions. Avoid publisher-name rules.

## Fixture and integration matrix
- NPR politics article with repeated hero photo, author portrait, photo credit: one lead, one portrait, caption preserved
- NPR Supreme Court article with small courthouse figure plus full-size original image: full-size lead retained, author portrait retained, figure retained
- NPR climate article with embedded audio player: no player HTML, headline once, all editorial paragraphs retained
- Article with large inline image and no metadata hero: inline image retained once
- Article with metadata hero but no inline images: hero shown
- Article with multiple gallery figures: all distinct figures retained in order
- Article with legitimately repeated photo (editorial choice): deduplication requires a shared source identity, not merely visual similarity
- Non-NPR publisher fixture with video/embed; prose and figures preserved
- Text-only article and print/PDF: no phantom images and no HTML artifacts

Tests must use frozen HTML fixtures with source attribution/URL, assert normalized output and rendered DOM where possible; separate tests for classification, metadata, and security guarantees. CI runs full tests and build. Mac acceptance checks Safari article view and print/PDF against the fixtures, not only version number.

## Scope and safety
Keep SSRF protections, HTTPS/public-host validation, redirect safeguards, byte/pixel caps, timeouts, endpoints, launch scripts, Cloudflare tunnel, and existing data untouched. No automatic merge into main. Increment visible version for each Mac candidate; supply complete local checkout, pull, test, build, restart, and Safari steps. Preserve readable content rather than suppressing evidence of parsing errors.
