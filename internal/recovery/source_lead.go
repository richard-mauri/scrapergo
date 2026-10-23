package recovery

import (
 "strings"
 "clean-article-reader/internal/media"
 "clean-article-reader/internal/sourcehtml"
 "clean-article-reader/internal/article"
)

// recoverOpeningSourceFigure preserves a leading editorial image that a
// text-oriented extractor dropped. All evidence is structural: figure/picture
// ancestry, article/main membership, document order, and exclusion of bylines,
// navigation and other non-editorial containers. An article-header image is
// eligible only with explicit lead/featured/hero-image evidence or figure/picture
// structure; a generic header avatar is not sufficient. No publisher identifiers.
func RecoverOpeningSourceFigure(a *article.Article, raw, sourceURL string) {
	if a == nil { return }
	root := sourcehtml.Parse(raw)
	for _, scope := range sourcehtml.Find(root, func(n *sourcehtml.Node) bool { return n.Tag == "article" }) {
		if RecoverFigureInScope(a, scope, sourceURL) { return }
	}
	for _, scope := range sourcehtml.Find(root, func(n *sourcehtml.Node) bool { return n.Tag == "main" }) {
		if RecoverFigureInScope(a, scope, sourceURL) { return }
	}
}

func RecoverFigureInScope(a *article.Article, scope *sourcehtml.Node, sourceURL string) bool {
	var candidate string
	var walk func(*sourcehtml.Node) bool
	sawProse := false
	walk = func(n *sourcehtml.Node) bool {
		if n != scope && (n.Tag == "aside" || n.Tag == "nav" || n.Tag == "footer") { return false }
		if n.Tag == "p" && len(strings.Fields(sourcehtml.Text(n))) >= 25 {
			sawProse = true
			return true
		}
		if n.Tag == "img" && !sawProse {
			ancestors := []media.Ancestor{}
			for p := n.Parent; p != nil && p != scope.Parent; p = p.Parent {
				ancestors = append(ancestors, media.Ancestor{Tag: p.Tag, Class: p.Attr["class"] + " " + p.Attr["id"]})
			}
			if media.OpeningFigureEligible(ancestors) && !sourcehtml.OpeningFigureNoisy(n, scope) {
				if src := sourcehtml.ImageURL(n, sourceURL); src != "" { candidate = src; return true }
			}
		}
		for _, child := range n.Kids { if walk(child) { return true } }
		return false
	}
	walk(scope)
	if candidate == "" { return false }
	return media.ReconcileOpeningLead(a, candidate, media.SourceIdentity)
}


