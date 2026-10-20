package main

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	readability "codeberg.org/readeck/go-readability/v2"
)

// extractWithReadability parses already-retrieved HTML; it makes no additional
// network requests. This preserves the existing SSRF protections and paywall gate.
func extractWithReadability(raw, source string) (Article, error) {
	u, err := url.Parse(source)
	if err != nil {
		return Article{}, err
	}
	parsed, parseErr := readability.FromReader(strings.NewReader(raw), u)
	if parseErr == nil {
		var output bytes.Buffer
		if err = parsed.RenderText(&output); err == nil {
			text := strings.TrimSpace(output.String())
			chunks := strings.Split(text, "\n\n")
			blocks := make([]Block, 0, len(chunks))
			count := 0
			for _, chunk := range chunks {
				line := normalize(chunk)
				if line == "" || sentenceAdRE.MatchString(line) {
					continue
				}
				blocks = append(blocks, Block{Kind: "p", Text: line})
				count += len(strings.Fields(line))
			}
			structuredRecovered := false
			if complete, ok := recoverStructuredDocument(raw, blocks, u); ok {
				blocks = complete
				count = countWords(blocks)
				structuredRecovered = true
			}
			// Restore missing opening paragraphs before accepting an otherwise plausible
			// Readability selection. Do not shift the remainder of the article.
			if !structuredRecovered {
				blocks = recoverLeadingParagraphs(raw, blocks, u)
				count = countWords(blocks)
			}
			if count >= 60 {
				title := parsed.Title()
				if title == "" {
					title = "Untitled article"
				}
				img := resolveImageURL(parsed.ImageURL(), u)
				if img == "" {
					img = resolveImageURL(meta(parseHTML(raw), "og:image", "twitter:image"), u)
				}
				var cleanHTML bytes.Buffer
				if parsed.RenderHTML(&cleanHTML) == nil {
					cleaned := cleanHTML.String()
					preserveSelectedHTMLLinks(blocks, cleaned, u)
					if !structuredRecovered {
						ordered := positionArticleFigures(blocks, cleaned, u)
						if strings.Contains(u.Path, "/gallery/") {
							ordered = supplementGalleryFigures(ordered, raw, u)
						}
						blocks = ordered
					}
				}
				if !structuredRecovered {
					blocks = recoverEditorialContinuity(raw, blocks, u)
					count = countWords(blocks)
				}
				if img == "" {
					for _, photo := range blocks {
						if photo.Kind == "img" {
							img = photo.URL
							break
						}
					}
				}
				for _, photo := range blocks {
					if photo.Kind == "img" && photo.URL == img {
						img = ""
						break
					}
				}
				return Article{Title: title, Byline: parsed.Byline(), Site: parsed.SiteName(), URL: source, Blocks: blocks, WordCount: count, ImageURL: img, Method: "Go Readability v2"}, nil
			}
		}
	}
	fallback, fallbackErr := extract(raw, source)
	if fallbackErr == nil {
		fallback.Method = "Legacy extractor fallback"
		return fallback, nil
	}
	if parseErr != nil {
		return fallback, fmt.Errorf("Readability: %v; legacy extractor: %w", parseErr, fallbackErr)
	}
	return fallback, fallbackErr
}

func resolveImageURL(raw string, base *url.URL) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	resolved := base.ResolveReference(u)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}
	if resolved.Hostname() == "" || resolved.User != nil {
		return ""
	}
	return resolved.String()
}

func articleImages(cleaned string, base *url.URL) []Block {
	root := parseHTML(cleaned)
	images := []Block{}
	seen := map[string]bool{}
	for _, n := range find(root, func(n *Node) bool { return n.Tag == "img" }) {
		src := n.Attr["src"]
		if src == "" {
			src = n.Attr["data-src"]
		}
		if src == "" {
			src = n.Attr["data-original"]
		}
		link := resolveImageURL(src, base)
		if link == "" || seen[link] {
			continue
		}
		seen[link] = true
		caption := ""
		for ancestor := n.Parent; ancestor != nil; ancestor = ancestor.Parent {
			if ancestor.Tag == "figure" {
				for _, c := range find(ancestor, func(x *Node) bool { return x.Tag == "figcaption" }) {
					caption = textOf(c)
					break
				}
				break
			}
			if ancestor.Tag == "body" {
				break
			}
		}
		if caption == "" {
			caption = normalize(n.Attr["alt"])
		}
		images = append(images, Block{Kind: "img", Text: caption, URL: link})
		if len(images) >= 12 {
			break
		}
	}
	return images
}

func attachInlineLinks(blocks []Block, cleaned string, base *url.URL) {
	root := parseHTML(cleaned)
	for _, anchor := range find(root, func(n *Node) bool { return n.Tag == "a" }) {
		label := normalize(textOf(anchor))
		if label == "" {
			continue
		}
		destination := resolveImageURL(anchor.Attr["href"], base)
		if destination == "" || !linkableDocument(destination) {
			continue
		}
		for i := range blocks {
			if blocks[i].Kind == "img" || !strings.Contains(blocks[i].Text, label) {
				continue
			}
			blocks[i].Links = append(blocks[i].Links, InlineLink{Text: label, URL: destination})
			break
		}
	}
}

func linkableDocument(destination string) bool {
	u, err := url.Parse(destination)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" || u.User != nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(u.Path[strings.LastIndex(u.Path, "/")+1:])) {
	case "":
		return true
	}
	name := strings.ToLower(u.Path)
	for _, suffix := range []string{".pdf", ".zip", ".gz", ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg", ".mp4", ".mp3", ".mov", ".avi", ".doc", ".docx", ".xls", ".xlsx"} {
		if strings.HasSuffix(name, suffix) {
			return false
		}
	}
	return true
}

func positionArticleFigures(blocks []Block, cleaned string, base *url.URL) []Block {
	root := parseHTML(cleaned)
	figures := articleImages(cleaned, base)
	if len(figures) == 0 {
		return blocks
	}
	placed := map[string]bool{}
	at := make(map[int][]Block)
	paragraph := 0
	for _, n := range find(root, func(n *Node) bool {
		return n.Tag == "p" || n.Tag == "h2" || n.Tag == "h3" || n.Tag == "h4" || n.Tag == "img"
	}) {
		if n.Tag != "img" {
			txt := normalize(textOf(n))
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
			for paragraph < len(blocks) && !strings.Contains(blocks[paragraph].Text, txt) && !strings.Contains(txt, blocks[paragraph].Text) {
				paragraph++
			}
			if paragraph < len(blocks) {
				paragraph++
			}
			continue
		}
		link := imageFromNode(n, base)
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
	out := make([]Block, 0, len(blocks)+len(figures))
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

func imageFromNode(n *Node, base *url.URL) string {
	for _, attr := range []string{"src", "data-src", "data-original"} {
		if link := resolveImageURL(n.Attr[attr], base); link != "" {
			return link
		}
	}
	return ""
}

func supplementGalleryFigures(blocks []Block, raw string, base *url.URL) []Block {
	root := parseHTML(raw)
	seen := map[string]bool{}
	for _, b := range blocks {
		if b.Kind == "img" {
			seen[b.URL] = true
		}
	}
	extra := []Block{}
	for _, fig := range find(root, func(n *Node) bool { return n.Tag == "figure" }) {
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
		imgNodes := find(fig, func(n *Node) bool { return n.Tag == "img" })
		if len(imgNodes) == 0 {
			continue
		}
		link := imageFromNode(imgNodes[0], base)
		if link == "" || seen[link] {
			continue
		}
		caption := ""
		for _, c := range find(fig, func(n *Node) bool { return n.Tag == "figcaption" }) {
			caption = normalize(textOf(c))
			break
		}
		if caption == "" {
			caption = normalize(imgNodes[0].Attr["alt"])
		}
		if caption == "" || len(caption) < 25 {
			continue
		}
		extra = append(extra, Block{Kind: "img", URL: link, Text: caption})
		seen[link] = true
		if len(extra) >= 25 {
			break
		}
	}
	return append(blocks, extra...)
}
