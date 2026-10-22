// Package extraction owns the handoff from source parsing to document assembly.
// Source evidence is kept separately from the canonical article document.
package extraction

import "clean-article-reader/internal/article"

// SourceEvidence records source observations without choosing how they should
// affect the final article. It intentionally has no HTTP, HTML, or UI dependencies.
type SourceEvidence struct {
 SourceURL string
 MetadataImageURL string
}

// Result is the boundary between the initial article extractor and the
// later document reconciliation stages.
type Result struct {
 Article article.Article
 Evidence SourceEvidence
}
