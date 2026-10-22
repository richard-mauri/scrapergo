package main

import (
 "strings"
 "clean-article-reader/internal/sourcehtml"
)

// recoverOpeningSourceFigure preserves a leading editorial image that a
// text-oriented extractor dropped. All evidence is structural: figure/picture
// ancestry, article/main membership, document order, and exclusion of bylines,
// navigation and other non-editorial containers. No publisher identifiers.
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
		if n != scope && (n.Tag == "aside" || n.Tag == "nav" || n.Tag == "footer" || n.Tag == "header") { return false }
		if n.Tag == "p" && len(strings.Fields(textOf(n))) >= 25 {
			sawProse = true
			return true
		}
		if n.Tag == "img" && !sawProse {
			eligible := false
			rejected := false
			for p := n.Parent; p != nil && p != scope.Parent; p = p.Parent {
				classes := strings.ToLower(p.Attr["class"] + " " + p.Attr["id"])
				if p.Tag == "figure" || p.Tag == "picture" { eligible = true }
				if p.Tag == "nav" || p.Tag == "aside" || p.Tag == "footer" || p.Tag == "header" ||
					strings.Contains(classes, "byline") || strings.Contains(classes, "author") ||
					strings.Contains(classes, "avatar") || strings.Contains(classes, "contributor") { rejected = true; break }
			}
			if eligible && !rejected && !sourcehtml.InNoisy(n, scope.Parent) {
				if src := articleImageURL(n, sourceURL); src != "" { candidate = src; return true }
			}
		}
		for _, child := range n.Kids { if walk(child) { return true } }
		return false
	}
	walk(scope)
	if candidate == "" { return false }
	for i := range a.Blocks {
		if a.Blocks[i].Kind != "img" { continue }
		if imageSourceIdentity(a.Blocks[i].URL) == imageSourceIdentity(candidate) {
			a.Blocks[i].Role = "lead"
			return true
		}
	}
	a.Blocks = append([]Block{{Kind: "img", URL: candidate, Role: "lead"}}, a.Blocks...)
	return true
}
