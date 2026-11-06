package extraction

import (
	"clean-article-reader/internal/article"
	"clean-article-reader/internal/sourcehtml"
	"net/url"
	"strings"
)

// RestoreLiveHeadings restores heading semantics lost when Readability's text
// output flattens live-update HTML into paragraph blocks. It changes neither
// text nor ordering, and only recognizes exact, unique source-backed headings
// inside an editorial article/main container on /live/ pages.
func RestoreLiveHeadings(raw string, blocks []article.Block, base *url.URL) []article.Block {
	if base == nil || !strings.Contains(base.Path, "/live/") || len(blocks) == 0 {
		return blocks
	}
	root := sourcehtml.Parse(raw)
	counts := map[string]int{}
	for _, h := range sourcehtml.Find(root, func(n *sourcehtml.Node) bool {
		return n.Tag == "h2" || n.Tag == "h3" || n.Tag == "h4"
	}) {
		if sourcehtml.InNoisy(h, nil) {
			continue
		}
		editorial := false
		for p := h.Parent; p != nil; p = p.Parent {
			if p.Tag == "article" || p.Tag == "main" {
				editorial = true
				break
			}
		}
		if !editorial {
			continue
		}
		key := sourcehtml.Normalize(sourcehtml.Text(h))
		if len(key) < 15 || len(key) > 220 {
			continue
		}
		counts[key]++
	}
	for i := range blocks {
		if blocks[i].Kind != "p" {
			continue
		}
		if counts[sourcehtml.Normalize(blocks[i].Text)] == 1 {
			blocks[i].Kind = "h3"
		}
	}
	return blocks
}
