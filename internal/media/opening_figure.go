package media

import "strings"

// OpeningFigureEligible evaluates source-image ancestry inside a confirmed
// article/main scope. Source traversal and noisy-container screening remain
// the responsibility of the HTML adapter. Closest ancestors come first.
func OpeningFigureEligible(ancestors []Ancestor) bool {
	eligible := false
	for _, ancestor := range ancestors {
		classes := strings.ToLower(ancestor.Class)
		if ancestor.Tag == "nav" || ancestor.Tag == "aside" || ancestor.Tag == "footer" ||
			strings.Contains(classes, "byline") || strings.Contains(classes, "author") ||
			strings.Contains(classes, "avatar") || strings.Contains(classes, "contributor") {
			return false
		}
		if ancestor.Tag == "figure" || ancestor.Tag == "picture" ||
			strings.Contains(classes, "featured-image") ||
			strings.Contains(classes, "hero-image") ||
			strings.Contains(classes, "lead-image") {
			eligible = true
		}
	}
	return eligible
}
