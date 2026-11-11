# Offline single-article duplicate provenance — Issue #119 / draft PR #120

Run this diagnostic **only against previously captured, authorized publisher HTML**. It invokes the established ScraperGo extraction and reader code without fetching a URL, without contacting any remote publisher, and without changing server behavior. The diagnostic prints final candidate block indices/types, a short nonreversible text digest, and the recovery stage counts. It does not print article content or persist raw HTML. It is not a production deduplication fix.

First, from the existing PR #120 checkout, capture each public article with the existing security-validated command. Use **distinct output paths**, as `htmlcapture` will not overwrite files:

```sh
cd ~/projects/scrapergo
go run ./cmd/htmlcapture \
  -url 'https://www.newscientist.com/article/2598048-a-true-paleo-diet-would-include-a-lot-more-insects/' \
  -out /tmp/pr120-paleo.html
go run ./cmd/readabilityprobe \
  -input /tmp/pr120-paleo.html \
  -url 'https://www.newscientist.com/article/2598048-a-true-paleo-diet-would-include-a-lot-more-insects/' \
  -out /tmp/pr120-paleo-readability
go run ./cmd/duplicateprobe \
  -input /tmp/pr120-paleo.html \
  -url 'https://www.newscientist.com/article/2598048-a-true-paleo-diet-would-include-a-lot-more-insects/'
```

Then independently inspect the original Ars example:

```sh
go run ./cmd/htmlcapture \
  -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/' \
  -out /tmp/pr120-ars-m6.html
go run ./cmd/readabilityprobe \
  -input /tmp/pr120-ars-m6.html \
  -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/' \
  -out /tmp/pr120-ars-m6-readability
go run ./cmd/duplicateprobe \
  -input /tmp/pr120-ars-m6.html \
  -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/'
```

If source fetching is rejected or returns incomplete challenge HTML, **stop and report it**; do not bypass access restrictions. `htmlcapture` may return source HTML different from a browser-rendered article; the probe does not emulate network-dependent PDF fallback or full HTTP page gating. Distinguish these offline outcomes from the full `/api/read` JSON.

You can paste just `duplicateprobe` stdout and `readabilityprobe`'s metadata `report.json` (not the HTML) to the PR reviewer. Private captures may contain copyrighted full-page text and must not be uploaded or committed indiscriminately. The block-index findings are a reproducible aid for inspecting the source locally, not a safe reason to remove text without checking legitimate occurrences.

If source snapshots differ from the earlier 100-URL run, record the discrepancy and retest with the same snapshot. Keep all source-backed fixes on this single draft PR #120.

## Ars M6: compare source and Readability HTML without printing article text

After a secure capture, run against the **same** saved HTML; do not recapture:

```sh
go run ./cmd/duplicateprobe \
  -input /tmp/pr120-ars-m6.html \
  -url 'https://arstechnica.com/gadgets/2026/11/review-apples-1999-m6-macbook-pro-feels-stuck-in-the-middle/' \
  -source-evidence
```

The optional `-source-evidence` report prints, for each pair of final duplicate-candidate blocks, how many source HTML and Readability-selected HTML **paragraph** or **figcaption** elements contain the shorter matched text. It uses the current private source HTML and the same Go Readability library but does not expose article text or write output files. Because this is a conservative substring presence test, zero counts can occur when HTML text serialization/inline formatting differs. Counts can also include repeated occurrences that are editorially legitimate. A single paragraph plus a caption in the original source is not the same problem as two independent editorial paragraphs. **Do not treat these counts as a deletion instruction.** Inspect raw source locally if the report is ambiguous.

### Source text that is not inside a paragraph

The `-source-evidence` report also displays **Source minimal tags** and **Readability minimal tags**. These counts search for the shortest matching source element across *all nonblocked tag types*, rather than only `p`/`figcaption`. The text is concatenated across inline HTML nodes without injecting spaces between tags, and parent containers are not double-counted when a child already contains the full match. For example `div:2` means two independently matched smallest `div` subtrees; it does **not** prove those copies are editorially identical in meaning. Script/style/navigation text is excluded under existing source parser blocked-tag policy.

If the source has no matching elements even with this broader scan, the unresolved explanations include text/entity normalization differences, publisher generated content, or Readability's own synthesis. Do not infer the cause from an unmatched count alone. Never upload or commit the full captured HTML without permission.
