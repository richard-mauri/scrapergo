package main

import (
	"net/url"
	"strings"
	"unicode"
)

// semanticKey compares complete paragraph text without depending on whitespace
// differences introduced by HTML layout or Readability's text renderer.
func semanticKey(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// preserveSelectedHTMLLinks keeps anchors in their original paragraph order.
// It annotates only Readability-selected blocks; it does not import other text,
// navigation, or image nodes into the article.
func preserveSelectedHTMLLinks(blocks []Block, cleaned string, base *url.URL) {
	root := parseHTML(cleaned)
	used := make(map[int]bool)
	for _, node := range find(root, func(n *Node) bool {
		return n.Tag == "p" || n.Tag == "li" || n.Tag == "h2" || n.Tag == "h3" || n.Tag == "h4" || n.Tag == "blockquote"
	}) {
		if inNoisy(node, nil) || len(find(node, func(n *Node) bool { return n.Tag == "a" })) == 0 {
			continue
		}
		key := semanticKey(textOf(node))
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
			if semanticKey(blocks[i].Text) != key {
				continue
			}
			used[i] = true
			blocks[i].Parts = parts
			blocks[i].Links = nil
			break
		}
	}
}
