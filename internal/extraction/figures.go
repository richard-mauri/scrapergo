// Package extraction owns source-HTML figure, gallery and document-link recovery.
package extraction

import (
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/sourcehtml"
 "clean-article-reader/internal/links"
 "net/url"
 "strings"
)

func ArticleImages(cleaned string, base *url.URL) []article.Block {
	root := sourcehtml.Parse(cleaned)
	images := []article.Block{}
	seen := map[string]bool{}
	for _, n := range sourcehtml.Find(root, func(n *sourcehtml.Node) bool { return n.Tag == "img" }) {
		src := n.Attr["src"]
		if src == "" {
			src = n.Attr["data-src"]
		}
		if src == "" {
			src = n.Attr["data-original"]
		}
		link := sourcehtml.ResolveURL(src, base)
		if link == "" || seen[link] {
			continue
		}
		seen[link] = true
		caption := ""
		for ancestor := n.Parent; ancestor != nil; ancestor = ancestor.Parent {
			if ancestor.Tag == "figure" {
				for _, c := range sourcehtml.Find(ancestor, func(x *sourcehtml.Node) bool { return x.Tag == "figcaption" }) {
					caption = sourcehtml.Text(c)
					break
				}
				break
			}
			if ancestor.Tag == "body" {
				break
			}
		}
		if caption == "" {
			caption = sourcehtml.Normalize(n.Attr["alt"])
		}
		images = append(images, article.Block{Kind: "img", Text: caption, URL: link})
		if len(images) >= 12 {
			break
		}
	}
	return images
}

func AttachInlineLinks(blocks []article.Block, cleaned string, base *url.URL) {
	root := sourcehtml.Parse(cleaned)
	for _, anchor := range sourcehtml.Find(root, func(n *sourcehtml.Node) bool { return n.Tag == "a" }) {
		label := sourcehtml.Normalize(sourcehtml.Text(anchor))
		if label == "" {
			continue
		}
		destination := sourcehtml.ResolveURL(anchor.Attr["href"], base)
		if destination == "" || !links.IsDocumentURL(destination) {
			continue
		}
		for i := range blocks {
			if blocks[i].Kind == "img" || !strings.Contains(blocks[i].Text, label) {
				continue
			}
			blocks[i].Links = append(blocks[i].Links, article.InlineLink{Text: label, URL: destination})
			break
		}
	}
}

func PositionArticleFigures(blocks []article.Block, cleaned string, base *url.URL) []article.Block {
	root := sourcehtml.Parse(cleaned)
	figures := ArticleImages(cleaned, base)
	if len(figures) == 0 {
		return blocks
	}
	placed := map[string]bool{}
	at := make(map[int][]article.Block)
	paragraph := 0
	for _, n := range sourcehtml.Find(root, func(n *sourcehtml.Node) bool {
		return n.Tag == "p" || n.Tag == "h2" || n.Tag == "h3" || n.Tag == "h4" || n.Tag == "img"
	}) {
		if n.Tag != "img" {
			txt := sourcehtml.Normalize(sourcehtml.Text(n))
			if txt == "" {
				continue
			}
			inFigure := false
			for x := n.Parent; x != nil; x = x.Parent {
				if x.Tag == "figure" || x.Tag == "figcaption" {
					inFigure = true
					break
				}
			}
			if inFigure {
				continue
			}
			// A source paragraph omitted or rewritten by Readability must not
			// consume the entire block cursor and push later images to the end.
			// Advance only when a retained prose block really matches.
			for next := paragraph; next < len(blocks); next++ {
				if blocks[next].Kind == "img" || blocks[next].Kind == "topic-link" {
					continue
				}
				candidate := sourcehtml.Normalize(blocks[next].Text)
				if candidate != "" && (strings.Contains(candidate, txt) || strings.Contains(txt, candidate)) {
					paragraph = next + 1
					break
				}
			}
			continue
		}
		link := ImageFromNode(n, base)
		if link == "" || placed[link] {
			continue
		}
		for _, f := range figures {
			if f.URL == link {
				at[paragraph] = append(at[paragraph], f)
				placed[link] = true
				break
			}
		}
	}
	out := make([]article.Block, 0, len(blocks)+len(figures))
	for i := 0; i <= len(blocks); i++ {
		out = append(out, at[i]...)
		if i < len(blocks) {
			out = append(out, blocks[i])
		}
	}
	for _, f := range figures {
		if !placed[f.URL] {
			out = append(out, f)
		}
	}
	return out
}

func ImageFromNode(n *sourcehtml.Node, base *url.URL) string {
	for _, attr := range []string{"src", "data-src", "data-original"} {
		if link := sourcehtml.ResolveURL(n.Attr[attr], base); link != "" {
			return link
		}
	}
	return ""
}

func SupplementGalleryFigures(blocks []article.Block, raw string, base *url.URL) []article.Block {
	root := sourcehtml.Parse(raw)
	seen := map[string]bool{}
	for _, b := range blocks {
		if b.Kind == "img" {
			seen[b.URL] = true
		}
	}
	extra := []article.Block{}
	for _, fig := range sourcehtml.Find(root, func(n *sourcehtml.Node) bool { return n.Tag == "figure" }) {
		allowed := false
		for p := fig.Parent; p != nil; p = p.Parent {
			if p.Tag == "article" || p.Tag == "main" || strings.Contains(strings.ToLower(p.Attr["class"]), "gallery") {
				allowed = true
				break
			}
		}
		if !allowed {
			continue
		}
		imgNodes := sourcehtml.Find(fig, func(n *sourcehtml.Node) bool { return n.Tag == "img" })
		if len(imgNodes) == 0 {
			continue
		}
		link := ImageFromNode(imgNodes[0], base)
		if link == "" || seen[link] {
			continue
		}
		caption := ""
		for _, c := range sourcehtml.Find(fig, func(n *sourcehtml.Node) bool { return n.Tag == "figcaption" }) {
			caption = sourcehtml.Normalize(sourcehtml.Text(c))
			break
		}
		if caption == "" {
			caption = sourcehtml.Normalize(imgNodes[0].Attr["alt"])
		}
		if caption == "" || len(caption) < 25 {
			continue
		}
		extra = append(extra, article.Block{Kind: "img", URL: link, Text: caption})
		seen[link] = true
		if len(extra) >= 25 {
			break
		}
	}
	return append(blocks, extra...)
}
