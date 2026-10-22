package main

import (
 "strings"
 "clean-article-reader/internal/media"
 "clean-article-reader/internal/sourcehtml"
)

// recoverOpeningSourceFigure preserves a leading editorial image that a
// text-oriented extractor dropped. All evidence is structural: figure/picture
// ancestry, article/main membership, document order, and exclusion of bylines,
// navigation and other non-editorial containers. An article-header image is
// eligible only with explicit lead/featured/hero-image evidence or figure/picture
// structure; a generic header avatar is not sufficient. No publisher identifiers.
func recoverOpeningSourceFigure(a *Article, raw, sourceURL string) {
	if a == nil { return }
	root := parseHTML(raw)
	for _, scope := range find(root, func(n *Node) bool { return n.Tag == "article" }) {
		if recoverFigureInScope(a, scope, sourceURL) { return }
	}
	for _, scope := range find(root, func(n *Node) bool { return n.Tag == "main" }) {
		if recoverFigureInScope(a, scope, sourceURL) { return }
	}
}

func recoverFigureInScope(a *Article, scope *Node, sourceURL string) bool {
	var candidate string
	var walk func(*Node) bool
	sawProse := false
	walk = func(n *Node) bool {
		if n != scope && (n.Tag == "aside" || n.Tag == "nav" || n.Tag == "footer") { return false }
		if n.Tag == "p" && len(strings.Fields(textOf(n))) >= 25 {
			sawProse = true
			return true
		}
		if n.Tag == "img" && !sawProse {
			ancestors := []media.Ancestor{}
			for p := n.Parent; p != nil && p != scope.Parent; p = p.Parent {
				ancestors = append(ancestors, media.Ancestor{Tag: p.Tag, Class: p.Attr["class"] + " " + p.Attr["id"]})
			}
			if media.OpeningFigureEligible(ancestors) && !sourcehtml.OpeningFigureNoisy(n, scope) {
				if src := articleImageURL(n, sourceURL); src != "" { candidate = src; return true }
			}
		}
		for _, child := range n.Kids { if walk(child) { return true } }
		return false
	}
	walk(scope)
	if candidate == "" { return false }
	return media.ReconcileOpeningLead(a, candidate, imageSourceIdentity)
}


