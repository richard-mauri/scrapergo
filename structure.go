package main

import (
 "clean-article-reader/internal/extraction"
 "clean-article-reader/internal/sourcehtml"
	"net/url"
)

// recoverStructuredDocument reads a single editorial subtree in DOM order.
// It is independent of hostnames, numbering schemes, and publisher templates.
// A repeated sequence of substantive headings is a strong sign that a generic
// text-only Readability selection may discard sibling sections.
func recoverStructuredDocument(raw string, reader []Block, base *url.URL) ([]Block, bool) {
	root := parseHTML(raw)
	var scope *Node
	var heads []*Node
	bestScore := 0
	for _, candidate := range find(root, func(n *Node) bool {
		return n.Tag == "article" || n.Tag == "main" || n.Tag == "section" || n.Tag == "div"
	}) {
		if sourcehtml.InNoisy(candidate, nil) {
			continue
		}
		hs := editorialHeadings(candidate)
		if len(hs) < 5 || len(hs) > 150 {
			continue
		}
		// Require substantive text outside headings, avoiding menus/catalogs.
		prose := 0
		for _, n := range find(candidate, func(n *Node) bool { return n.Tag == "p" || n.Tag == "li" }) {
			if !sourcehtml.InNoisy(n, candidate) && !nestedProse(n, candidate) && len(normalize(textOf(n))) >= 35 {
				prose++
			}
		}
		if prose < len(hs)/2 {
			continue
		}
		score := len(hs)*10 + prose
		if score > bestScore || (score == bestScore && scope != nil && nodeDepth(candidate) > nodeDepth(scope)) {
			bestScore, scope, heads = score, candidate, hs
		}
	}
	if scope == nil {
		return nil, false
	}
	seenHeadings := make(map[string]bool)
	for _, b := range reader {
		if b.Kind == "h2" || b.Kind == "h3" || b.Kind == "h4" || b.Kind == "p" {
			seenHeadings[normalize(b.Text)] = true
		}
	}
	missing := 0
	for _, h := range heads {
		if !seenHeadings[normalize(textOf(h))] {
			missing++
		}
	}
	// Use structure when Readability has demonstrably lost sections, or when
	// its flattened output lacks semantics in a densely structured article.
	if missing < 2 && len(reader) > 0 && countWords(reader) >= 100 {
		return nil, false
	}
	photos := serializedCaptionImages(raw, base) // serialized image-caption pairs
	used := map[string]bool{}
	blocks := make([]Block, 0, len(heads)*5)
	var walk func(*Node)
	walk = func(n *Node) {
		if sourcehtml.InNoisy(n, scope) || promotionalRegion(n) || promotionalAncestor(n, scope) {
			return
		}
		if title, links, ok := contextualLinkSection(n, base); ok {
			blocks = append(blocks, Block{Kind: "topic-title", Text: title})
			for _, link := range links {
				blocks = append(blocks, link)
			}
			return
		}
		if navigationCluster(n) {
			return
		}
		switch n.Tag {
		case "p", "li", "h2", "h3", "h4", "blockquote", "figcaption":
			if nestedProse(n, scope) {
				return
			}
			b := semanticTextBlock(n, base)
			if b.Text == "" || extraction.IsAdvertisingText(b.Text) {
				return
			}
			if url := photos[normalize(b.Text)]; url != "" && !used[url] {
				blocks = append(blocks, Block{Kind: "img", Text: b.Text, URL: url})
				used[url] = true
				return
			}
			blocks = append(blocks, b)
			return
		case "img":
			url := imageFromNode(n, base)
			if url == "" || used[url] {
				return
			}
			caption := normalize(n.Attr["alt"])
			for a := n.Parent; a != nil && a != scope; a = a.Parent {
				if a.Tag == "figure" {
					for _, c := range find(a, func(x *Node) bool { return x.Tag == "figcaption" }) {
						caption = normalize(textOf(c))
						break
					}
					break
				}
			}
			blocks = append(blocks, Block{Kind: "img", Text: caption, URL: url})
			used[url] = true
			return
		case "script", "style", "noscript", "iframe", "svg", "nav", "aside", "footer", "header", "button", "form":
			return
		}
		for _, c := range n.Kids {
			walk(c)
		}
	}
	walk(scope)
	if len(blocks) < len(heads)+3 || countWords(blocks) < 80 {
		return nil, false
	}
	// Verify that every structural section is retained exactly once.
	have := map[string]int{}
	for _, b := range blocks {
		if b.Kind == "h2" || b.Kind == "h3" || b.Kind == "h4" {
			have[normalize(b.Text)]++
		}
	}
	for _, h := range heads {
		if have[normalize(textOf(h))] != 1 {
			return nil, false
		}
	}
	// Keep unrelated recommendations out of the document even when a site's
	// layout separates their image, caption, and heading into sibling nodes.
	blocks = trimTrailingRecommendation(blocks)
	// Contextual indexes are sometimes siblings of the editorial subtree rather
	// than children. Recover the nearest qualifying index without widening scope.
	if !hasContextualIndex(blocks) {
		if title, links, ok := findNearbyContextualIndex(root, scope, base); ok {
			blocks = append(blocks, Block{Kind: "topic-title", Text: title})
			blocks = append(blocks, links...)
		}
	}
	return blocks, true
}

func editorialHeadings(n *Node) []*Node {
	var out []*Node
	for _, h := range find(n, func(x *Node) bool { return x.Tag == "h2" || x.Tag == "h3" }) {
		if sourcehtml.InNoisy(h, n) || inNavigationCluster(h, n) || promotionalAncestor(h, n) || topicHeading(h) {
			continue
		}
		title := normalize(textOf(h))
		if len(title) < 5 || len(title) > 160 {
			continue
		}
		out = append(out, h)
	}
	return out
}

// This conservative detector deliberately declines ambiguous documents.
