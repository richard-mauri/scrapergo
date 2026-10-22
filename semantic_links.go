package main

import (
 "clean-article-reader/internal/sourcehtml"
 "clean-article-reader/internal/links"
	"net/url"
)

// preserveSelectedHTMLLinks keeps anchors in their original paragraph order.
// It annotates only Readability-selected blocks; it does not import other text,
// navigation, or image nodes into the article.
func preserveSelectedHTMLLinks(blocks []Block, cleaned string, base *url.URL) {
	root := parseHTML(cleaned)
	used := make(map[int]bool)
	for _, node := range find(root, func(n *Node) bool {
		return n.Tag == "p" || n.Tag == "li" || n.Tag == "h2" || n.Tag == "h3" || n.Tag == "h4" || n.Tag == "blockquote"
	}) {
		if sourcehtml.InNoisy(node, nil) || len(find(node, func(n *Node) bool { return n.Tag == "a" })) == 0 {
			continue
		}
		key := links.SemanticKey(textOf(node))
		if key == "" {
			continue
		}
		parts := inlineTextParts(node, base.String())
		hasLink := false
		for _, p := range parts {
			if p.URL != "" {
				hasLink = true
				break
			}
		}
		if !hasLink {
			continue
		}
		for i := range blocks {
			if used[i] || blocks[i].Kind == "img" || blocks[i].Kind == "topic-link" {
				continue
			}
			if links.SemanticKey(blocks[i].Text) != key {
				continue
			}
			used[i] = true
			blocks[i].Parts = parts
			blocks[i].Links = nil
			break
		}
	}
}
