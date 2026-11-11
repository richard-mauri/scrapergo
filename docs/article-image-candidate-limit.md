# Configurable article-image candidate limit — Issue #129

This change **does not improve editorial-image selection by itself**. It only replaces the fixed first-12 bound with a named, validated extraction configuration. The existing algorithm still collects valid unique image URLs in selected-HTML DOM order, preserving current caption fallback, ordering and URL safety. Do not infer that all extra candidates are photographs that should appear in the reader.

## Current setting

| Property | Value |
| --- | --- |
| Environment variable | `SCRAPERGO_ARTICLE_IMAGE_CANDIDATE_LIMIT` |
| Owned by | `internal/extraction.FigureSelectionConfig` |
| Default when unset | **12** unique valid image URLs (unchanged) |
| Allowed range | **1–200** candidates; both endpoints inclusive |
| Units | Number of distinct resolved image URLs considered for insertion |
| Zero | **Invalid**; not unlimited or disabled |
| Invalid or empty override | Explicit extraction error; never silently ignored |
| Override precedence | Process environment if present, otherwise compiled default |
| Change timing | Set before launching/restarting ScraperGo; no live-reload control |
| Security/resource bound | Maximum 200 bounds allocation and processing, not an editorial recommendation |

For an opt-in local experimental run, set the variable in the **environment of the ScraperGo process**, for example:

```bash
export SCRAPERGO_ARTICLE_IMAGE_CANDIDATE_LIMIT=14
./scrapergo.sh
```

Restore the normal default by unsetting the variable **and restarting the process**:

```bash
unset SCRAPERGO_ARTICLE_IMAGE_CANDIDATE_LIMIT
./scrapergo.sh
```

Do not run both launch commands in a row without stopping the running service using the existing owner-managed lifecycle. In ordinary operation, leave the variable unset for exact default compatibility. This is a **process setting**, not a Safari UI preference and not a setting stored in the user's account.

## Code ownership and configuration

`NewFigureSelectionConfig(limit)` validates bounds, `DefaultFigureSelectionConfig()` preserves 12, and `FigureSelectionConfigFromEnvironment()` performs an owner-controlled override lookup. `ArticleImagesWithConfig` and `PositionArticleFiguresWithConfig` receive config explicitly; the old direct `ArticleImages` and `PositionArticleFigures` APIs keep their default behavior for existing code/tests. The actual `ExtractWithReadabilityTraced` production pipeline reads the configuration and injects it into `PositionArticleFiguresWithConfig`; the same path is used by `ExtractWithReadability` and bootstrap. The extraction path makes no network requests and does not alter fetcher/access restrictions.

The candidate cap is **not** shared with `internal/recovery/source_gallery.go`'s independent gallery 2–12 structural rule or `SupplementGalleryFigures`' 25-image cap. Each must be considered separately under Issue #127 policy.

## Tests and acceptance

Synthetic test coverage checks default 12 from 14 explicit figures; direct configured selection at 1, 11, 12, 13, 14 and 15; image ordering and explicit captions; unsafe image URL/duplicate handling; invalid, empty and out-of-range environment values; and actual Readability extraction with 12 vs 14 candidate limits. The unchanged default must pass existing full repo Go tests. Any failure in synthetic production-path tests must be investigated rather than relaxing assertions solely to pass CI.

**Not yet validated:** owner Mac application build/launch, Safari/Print/PDF, multi-publisher real HTML, resource/performance at 200 images, and editorial appropriateness of newly admitted images. The unchanged 12 limit is the production default until the owner separately authorizes a change. Do not merge/deploy/tag/release without explicit approval.

Related: Issue #127 / draft PR #128 requires all such behavioral parameters to be named and tunable; Issue #125 / draft PR #126 explores generalized structural extraction and actual image ownership, including why the first-N strategy is inadequate.
