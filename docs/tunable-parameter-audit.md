# Tunable behavioral parameter audit — Issue #127

**Status:** Initial, non-exhaustive inspection of current merged `main` (baseline `2832ac7cde154af2910042cd030feb67f9f905d3`), November 11, 2026. Documentation only. **No number in this inventory has been replaced, exposed, or changed yet.** This inventory prioritizes semantics-affecting limits in extraction/media/recovery and must be expanded with additional source inspection and tests. Line numbers refer to that baseline and may move.

## Owner's binding design constraint

Behavioral numeric thresholds, cutoffs, scoring weights, retry/count limits, timeouts, quotas and byte/text limits need named tunable parameters, centralized defaults, explicit range/unit/override contracts and below/at/above boundary tests. Fixed mathematical/protocol invariants and structural indices are not arbitrary knobs; document why they are fixed. Security limits must not become unvalidated public configuration. No blanket global integer with unrelated meanings.

## First-pass findings (confirmed source locations)

| Source and line on baseline | Embedded literal / behavior | Parameter candidate | Priority and risk |
| --- | --- | --- | --- |
| `internal/extraction/figures.go:46` | `len(images) >= 12`: first 12 valid unique images from selected HTML are considered by `ArticleImages` | `ArticleImageCandidateLimit` (count; current default 12) | **High:** prevents later candidates. Needs editorial selection policy, not only a higher cap |
| `internal/extraction/figures.go:188,193` | supplementary gallery image requires caption length ≥25, then at most 25 additions | `GallerySupplementMinCaptionCharacters` (characters, 25); `GallerySupplementImageLimit` (count, 25) | High: may drop valid editorial photos/captions |
| `internal/recovery/source_gallery.go:40,64` | gallery layout accepts 2–12 image branches; later URL candidate groups capped at 12 | `GalleryMinDistinctImages` (count, 2); `GalleryMaxDistinctImages` (count, 12), with explicit distinction between branch and URL counts if semantic policy diverges | High: independent of `ArticleImageCandidateLimit`; a *separate* mechanism |
| `internal/extraction/readability_pipeline.go:71` | 60-word threshold for proceeding through selected recovery pipeline | `ReadabilityMinWordsForRecovery` (words, 60) | High: shifts extractor/fallback behavior |
| `internal/extraction/partial_excerpt.go:78,88` | partial preview max 32 blocks; selected acceptance 30–59 words | `PartialPreviewMaxBlocks` (count, 32), `PartialPreviewMinWords` (words, 30), `PartialPreviewBelowFullArticleWords` (words, 60) | High: access/restriction and article completeness, security/access-policy review required |
| `internal/extraction/navigation.go:21,27,31,125` | link-count 5–250, label length ≤65, short-link ratio 80%, max 80 selected links | `NavigationMinLinks`, `NavigationMaxLinks`, `NavigationMaxLabelCharacters`, `NavigationShortLinkPercent`, `NavigationSelectedLinkLimit` | Medium/high: avoid misclassifying landing pages as articles |
| `internal/extraction/heading_prose.go:35,52,68,84` | ≥3 anchors; text length 70–5000; search distance ≤75 nodes; max 10 recovered paragraphs per section | `HeadingMinAnchors`, `HeadingProseMinCharacters`, `HeadingProseMaxCharacters`, `HeadingProseMaxNodeDistance`, `HeadingProseMaxParagraphs` | High: generality/omitted article content |
| `internal/extraction/continuity_recovery.go:23,99,148,282` | ≥65-character anchors; short 25-character exclusion; ≥75-character source paragraphs; >80 tail paragraphs rejected | `ContinuityMinAnchorCharacters`, `ContinuityMinParagraphCharacters`, `ContinuityMinSourceParagraphCharacters`, `ContinuityMaxTailParagraphs` | Medium/high: interacting recovery rules; check other instances before refactor |
| `internal/extraction/figure_caption_credit.go:66,81,199` | caption text lengths 30 and 24/5, candidate matching cardinality 1–12 | Separate documented `CaptionReconciliation*` length / relationship thresholds, with unit (runes vs bytes), plus **investigate** whether 1–12 is a fixed cardinality check or tunable heuristic | High: false caption ownership; **do not conflate this 12 with image candidate cap** |
| `internal/extraction/legacy.go:146,148,154,156,159,162` | scoring contribution capped at 750; score boosts +50, +20 per paragraph, +1700, +300 etc. | Named `LegacySelectionWeights` with weights and thresholds | Medium: ranking behavior currently opaque |
| `internal/extraction/tables.go:23,24` | declared HTML `colspan` / `rowspan` capped at 100 | `MaxTableSpan` or explicit *bounded structural safety invariant* after security review | Medium: validate markup resource limits before making tunable |
| `internal/recovery/source_photographs.go:43,64` | alt text truncated at 180 bytes; proximity paragraph threshold 35 characters | `PhotoAltMaxBytes`, `PhotoAnchorMinCharacters` | Medium: safety/UTF-8 and loss of content; design should prefer DOM relationships over proximity |
| `internal/recovery/source_lead.go:33` | 25-word prose threshold to stop searching for opening media | `OpeningLeadStopAtProseWords` | Medium: can misclassify lead media |
| `internal/media/author_portraits.go:17` | 90-character paragraph threshold marks article as started | `AuthorPortraitEditorialStartCharacters` | Medium: can change contributor image classification |
| `internal/extraction/structure.go:27,40,62,125` | 5–150 candidate headings; heading score*10; 100-word / 80-word gates | Separate `StructuralSelection*` thresholds and weights | Medium: article scope/scoring, cross-publisher fixtures needed |
| `internal/extraction/duplicate_trace_counts.go:22,25,28` | 24/32/30-rune duplicate-screening lengths | `DuplicateTrace*` thresholds **only if they influence behavior**; distinguish diagnostic metric from runtime filters | Low/medium: verify if observational only before exposing controls |

Values above are **current observed literals**, *not recommended future defaults* or a complete enumeration. They may be part of interacting logic and cannot safely be replaced by mechanical search/replace.

## Implementation plan / configuration contract (future change)

Define subsystem-owned types (e.g. extraction selection, media recovery, page classification, navigation) with centralized defaults and validation; inject into the pipeline rather than referencing package-wide mutable globals. Validate override precedence and explicit invalid-value behavior, including zero/unlimited semantics, minimums/maximums and overflow. Choose an operational override interface only with owner approval after reviewing restart, security and user-facing settings needs. Avoid globally enabling every threshold as an anonymous unrestricted end-user setting.

For the first production change, prioritize `ArticleImageCandidateLimit` **alongside** redesign of source-backed editorial-media selection. Test at least 13 genuine editorial images with distinct captions and order, normal articles with fewer images, responsive alternate URLs, hidden/advertising icons, cross-publisher captures, final canonical reader output and Safari Print/PDF. Preserve the current default while parameterizing in a *separate implementation PR*; any proposal to change the default or select additional images requires evidence and owner authorization.

## Pending

- Broaden audit to every behavioral limit in extraction, media, recovery, reader, access policy, network and diagnostics (not just representative Go numeric comparisons).
- Agree on parameter names, types, grouping, override mechanism, validation and whether configuration is runtime reloadable.
- Check tests for current implicit limits; record default-compatibility evidence.
- Introduce static review or lint checks for new unexplained numeric literals if maintainable without false positives.
- Do not merge Issue #125's experimental structural converter into this documentation policy PR; no production changes are authorized here.
