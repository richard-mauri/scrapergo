package extraction

import (
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/sourcehtml"
	"net/url"
	"strings"
)

// recoverLeadingParagraphs restores substantive introduction paragraphs that
// Readability discarded before the first paragraph it retained. It never
// fetches another URL and operates only on the already downloaded HTML.
// The anchor must occur within an editorial article/main subtree; this
// avoids turning page-wide teaser or navigation paragraphs into article text.
func RecoverLeadingParagraphs(raw string, blocks []article.Block, base *url.URL, inlineParts func(*sourcehtml.Node,string) []article.TextPart) []article.Block {
	if len(blocks) == 0 || base == nil {
		return blocks
	}
	first := ""
	for _, b := range blocks {
		if b.Kind == "p" && len(sourcehtml.Normalize(b.Text)) >= 65 {
			first = sourcehtml.Normalize(b.Text)
			break
		}
	}
	if first == "" {
		return blocks
	}

	root := sourcehtml.Parse(raw)
	var anchor *sourcehtml.Node
	// Prefer a paragraph whose text begins with the first retained paragraph,
	// avoiding coincidental matches in summaries or recommendation widgets.
	for _, p := range sourcehtml.Find(root, func(n *sourcehtml.Node) bool { return n.Tag == "p" }) {
		t := sourcehtml.Normalize(sourcehtml.Text(p))
		if len(t) >= 65 && (strings.HasPrefix(first, t) || strings.HasPrefix(t, first)) && !sourcehtml.InNoisy(p, nil) {
			anchor = p
			break
		}
	}
	if anchor == nil {
		return blocks
	}

	var scope *sourcehtml.Node
	for n := anchor.Parent; n != nil; n = n.Parent {
		if n.Tag == "article" {
			scope = n
			break
		}
	}
	if scope == nil {
		for n := anchor.Parent; n != nil; n = n.Parent {
			if n.Tag == "main" {
				scope = n
				break
			}
		}
	}
	if scope == nil {
		return blocks
	}

	paragraphs := sourcehtml.Find(scope, func(n *sourcehtml.Node) bool { return n.Tag == "p" })
	var intro []article.Block
	for _, p := range paragraphs {
		if p == anchor {
			break
		}
		if sourcehtml.InNoisy(p, scope) || PromotionalAncestor(p, scope) || PromotionalRegion(p) || NestedProse(p, scope) {
			continue
		}
		t := sourcehtml.Normalize(sourcehtml.Text(p))
		if len(t) < 25 || IsAdvertisingText(t) {
			continue
		}
		duplicated := false
		for _, b := range blocks {
			current := sourcehtml.Normalize(b.Text)
			if current == t || (len(current) >= 65 && (strings.Contains(current, t) || strings.Contains(t, current))) {
				duplicated = true
				break
			}
		}
		if duplicated {
			continue
		}
		// Avoid accidentally importing a whole unrelated section. An intro
		// spanning more than 16 distinct paragraphs is too ambiguous.
		if len(intro) >= 16 {
			return blocks
		}
		intro = append(intro, article.Block{Kind: "p", Text: t, Parts: inlineParts(p, base.String())})
	}
	if len(intro) == 0 {
		return blocks
	}
	return append(intro, blocks...)
}
