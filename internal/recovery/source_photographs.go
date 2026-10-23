package recovery

import (
 "strings"
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/sourcehtml"
)

// recoverArticlePhotographs adds photographs from article-associated image wrappers
// that a text-oriented article selector may have excluded. Only images in
// explicit editorial wrappers are eligible: navigation and reaction icons are
// intentionally not treated as article photographs.
func RecoverArticlePhotographs(a *article.Article, raw, baseURL string) {
	root := sourcehtml.Parse(raw)
	existing := map[string]bool{}
	for _, b := range a.Blocks {
		if b.Kind == "img" {
			existing[b.URL] = true
		}
	}
	type candidate struct{ url, alt, near string }
	var found []candidate
	for _, n := range sourcehtml.Find(root, func(n *sourcehtml.Node) bool { return n.Tag == "img" }) {
		if sourcehtml.InNoisy(n, nil) {
			continue
		}
		editorial := false
		for p := n.Parent; p != nil; p = p.Parent {
			cls := strings.ToLower(p.Attr["class"] + " " + p.Attr["id"])
			if p.Tag == "figure" || strings.Contains(cls, "full-width-image") || strings.Contains(cls, "article-content") || strings.Contains(cls, "article-body") {
				editorial = true
				break
			}
		}
		if !editorial {
			continue
		}
		u := sourcehtml.ImageURL(n, baseURL)
		if u == "" || existing[u] {
			continue
		}
		alt := sourcehtml.Normalize(n.Attr["alt"])
		if len(alt) > 180 {
			alt = alt[:180]
		}
		// Find the next nearby text paragraph in document order to anchor the
		// picture, while keeping the original article's paragraphs unchanged.
		near := ""
		for p := n.Parent; p != nil && near == ""; p = p.Parent {
			if p.Parent == nil {
				break
			}
			following := false
			for _, sib := range p.Parent.Kids {
				if sib == p {
					following = true
					continue
				}
				if !following {
					continue
				}
				for _, para := range sourcehtml.Find(sib, func(x *sourcehtml.Node) bool { return x.Tag == "p" }) {
					t := textOf(para)
					if len(t) >= 35 {
						near = t
						break
					}
				}
				if near != "" {
					break
				}
			}
		}
		found = append(found, candidate{u, alt, near})
		existing[u] = true
	}
	if len(found) == 0 {
		return
	}
	original := a.Blocks
	before := make(map[int][]article.Block)
	tail := []article.article.Block{}
	lastAnchor := -1
	for _, f := range found {
		img := article.Block{Kind: "img", URL: f.url, Text: f.alt}
		anchor := -1
		if f.near != "" {
			for i, b := range original {
				if b.Kind == "p" && (strings.Contains(b.Text, f.near) || strings.Contains(f.near, b.Text)) {
					anchor = i
					break
				}
			}
		}
		if anchor < 0 && lastAnchor >= 0 {
			anchor = lastAnchor + 1
		}
		if anchor >= 0 && anchor <= len(original) {
			before[anchor] = append(before[anchor], img)
			lastAnchor = anchor
		} else {
			tail = append(tail, img)
		}
	}
	merged := make([]article.Block, 0, len(original)+len(found))
	for i, b := range original {
		merged = append(merged, before[i]...)
		merged = append(merged, b)
	}
	merged = append(merged, before[len(original)]...)
	a.Blocks = append(merged, tail...)
}

