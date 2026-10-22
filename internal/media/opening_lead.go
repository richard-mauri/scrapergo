package media

import "clean-article-reader/internal/article"

// ReconcileOpeningLead preserves an already rendered rendition of the same
// opening image or prepends a missing image without discarding later photos.
// The caller supplies source URL resolution and rendition-aware identity.
func ReconcileOpeningLead(a *article.Article, candidate string, identity func(string) string) bool {
	if a == nil || candidate == "" || identity == nil {
		return false
	}
	for i := range a.Blocks {
		if a.Blocks[i].Kind != "img" {
			continue
		}
		if identity(a.Blocks[i].URL) == identity(candidate) {
			a.Blocks[i].Role = "lead"
			return true
		}
	}
	a.Blocks = append([]article.Block{{Kind: "img", URL: candidate, Role: "lead"}}, a.Blocks...)
	return true
}
