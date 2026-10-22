package sourcehtml

import "regexp"

// Shared source-noise classification used by source analysis and page gates.
var promotRE = regexp.MustCompile(`(?i)(advertis|sponsor|promo|newsletter|signup|sign-up|subscribe|paywall|related|recommended|trending|most-read|share[-_ ]|social|ticker|stock[-_ ]?quote|market[-_ ]?data|author[-_ ]?bio|affiliate|outbrain|taboola|comments|breadcrumb|sidebar|sticky|modal|popup|cookie|footer|navigation|nav-menu|more-content|continue-reading|biggest-winner|biggest-loser)`)

func Noisy(n *Node) bool {
	if BlockedTags[n.Tag] {
		return true
	}
	s := n.Attr["class"] + " " + n.Attr["id"] + " " + n.Attr["role"] + " " + n.Attr["aria-label"]
	return promotRE.MatchString(s)
}
func InNoisy(n, stop *Node) bool {
	for p := n; p != nil && p != stop; p = p.Parent {
		if Noisy(p) {
			return true
		}
	}
	return false
}

// OpeningFigureNoisy applies ordinary source noise rules to an opening image,
// except that an article's own header is structural rather than automatically
// disqualifying. Other noisy ancestors, including bylines, remain blocked.
// Traversal stops at the parent of scope, matching source recovery boundaries.
func OpeningFigureNoisy(n, scope *Node) bool {
	if n == nil || scope == nil {
		return true
	}
	for p := n; p != nil && p != scope.Parent; p = p.Parent {
		if p.Tag == "header" && scope.Tag == "article" {
			continue
		}
		if Noisy(p) {
			return true
		}
	}
	return false
}
