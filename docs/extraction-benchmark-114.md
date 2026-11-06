# Issue #114 — extraction benchmark baseline and equivalence gate

**Scope:** Reproducible measurement and approval criteria for future changes, not production pipeline modification. [Issue #114](https://github.com/richard-mauri/scrapergo/issues/114) / [draft PR #115](https://github.com/richard-mauri/scrapergo/pull/115). The accepted CBS headline/time rendering from merged [PR #109](https://github.com/richard-mauri/scrapergo/pull/109) must remain unchanged.

## What is measured

`internal/extraction/readability_pipeline_benchmark_test.go` defines deterministic synthetic fixtures with 18 substantive paragraphs. `BenchmarkReadabilityExtractionOrdinary` represents a conventional multi-section article; `BenchmarkReadabilityExtractionMultiUpdate` adds three `<time>` + `<h2>` update boundaries. Both call the real `ExtractWithReadability` with its existing legacy fallback and link hooks. Each performs an **untimed** setup check that the fixture entered the intended Go Readability v2 branch, has at least 60 words and has a substantial block count. The timing loop measures the complete Readability extraction including internal recovery stages. It does **not** measure HTTP retrieval, pipeline composition, browser rendering, Safari, or a live publisher's DOM.

The preexisting `BenchmarkPreserveReadabilityStructure` in `internal/extraction/readability_structure_characterization_test.go` measures only the heading/timestamp overlay with 25 synthetic updates. This lets a future candidate distinguish overlay-specific cost from end-to-end extraction cost. Both are microbenchmarks with fixture limitations.

## Repeatable measurement commands

From the repository root on the same machine, Go version and environment for both baseline and candidate. The normal `./devbuild.sh` executes `go test ./...` and then `bash ./benchmark.sh --smoke` before building either architecture; the smoke checks fail the build only if the benchmark cannot execute correctly, **not** when execution times vary. Run `./benchmark.sh` independently for five measured runs:

```sh
go test -count=1 ./...
./benchmark.sh
```

For comparative work, save output from baseline and candidate under `/tmp` (not the repository):

```sh
./benchmark.sh > /tmp/scrapergo-baseline-bench.txt
# After checking out a separately approved candidate, run exactly the same command:
./benchmark.sh > /tmp/scrapergo-candidate-bench.txt
```

For reliable comparison, use identical machine conditions and the same Go toolchain and CPU architecture. A prebuilt `benchstat` binary can summarize the two files **if already available**; do not require Homebrew source compilation just for this. Inspect `ns/op`, `B/op`, `allocs/op`, variance and sample counts. `B/op` reports allocations, **not** peak memory. Longer/shorter timings on one synthetic fixture do not prove real-world performance. On an older Intel Mac with limited RAM, avoid concurrent heavy builds or benchmarks; benchmark output is intentionally opt-in and not part of routine CI.

**Measurements obtained:** None yet. A previous temporary CI attempt in PR #113 failed before jobs ran and its workflow changes were reverted; no timings can be inferred from it. Do not claim speedup or regression without actual results.

## Output-equivalence decision gate (before any production refactor)

Before proposing a caching layer, shared parsed DOM, stage reorder or direct HTML-to-block conversion, collect baseline and candidate `article.Article` results from the same permitted inputs (all synthetic tests plus properly permissioned representative publisher captures). Compare exact relevant fields, and document intentional differences for owner approval:

| Input / scenario | Required equivalence |
| --- | --- |
| Ordinary article with multiple sections | Title, byline, editorial paragraphs, source order and heading levels preserved |
| Live updates with time tags and headings | No attached timestamps, missing headings or interleaved updates; retain original prose |
| Duplicate or ambiguous headings/time-like prose | No invented headings, split timestamps, or duplicate sections |
| Links and rich inline text | Inline `Parts`, URL targets, emphasis, punctuation and source order unchanged |
| Editorial images, portraits, captions, galleries | Source image roles, identity, URLs, captions, deduplication and order unchanged |
| Tables, document/PDF links | Cells, spans, headers, linked content and text serialization unchanged |
| Partial, blocked, paywalled, malformed, or short input | Same access-policy decisions, fallback path, error/preview and word-count behavior |
| Index/classification and duplicate-live-update filtering | Preserve article/index recognition and long-repeat cleanup without losing legitimate updates |
| Reader, print/PDF and Safari | Owner checks UI formatting, navigation, photos, source links and print/export; CI cannot substitute |
| Retrieval boundaries | No new network requests in extraction or bypass of SSRF, redirect or access restrictions |

The earlier [Issue #110 audit](extraction-cleanup-110.md) identifies live responsibilities that cannot be deleted merely because they overlap in function names. New benchmarks establish repeatable *infrastructure*, **not** a design decision. For a future refactor, first gather benchmark data, output equivalence evidence, and a rollback plan; then obtain separate owner authorization. No publisher-specific heuristics or silent changes to the preserved CBS behavior.

## Verification / handoff

CI on the latest PR head must pass `go test -count=1 ./...` and `go build ./...`. The full five-repeat benchmark is opt-in; the short smoke benchmark is now part of `./devbuild.sh` and is not part of plain `go test ./...` or the GitHub Actions workflow. Synthetic CI fixture passes do not imply a real publisher or Safari was tested. This PR is intentionally test/documentation-only and requires no local app restart for review. Do not merge, tag or release without the owner's explicit approval.
