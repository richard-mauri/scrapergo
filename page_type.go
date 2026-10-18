package main

import (
	"net/url"
	"strings"
)

// pageIsIndex recognizes link-driven collections instead of trusting Readability's
// word count. This classifier is deliberately independent of hostnames and paths.
// Ambiguous documents are passed to the normal article extractor.
func pageIsIndex(raw string) bool {
	root := parseHTML(raw)
	scope := root
	for _, n := range find(root, func(n *Node) bool { return n.Tag == "main" }) {
		if !inNoisy(n, nil) {
			scope = n
			break
		}
	}
	links, headlineLinks, prose := 0, 0, 0
	targets := map[string]bool{}
	for _, n := range find(scope, func(n *Node) bool { return n.Tag == "a" || n.Tag == "p" }) {
		if inNoisy(n, scope) {
			continue
		}
		switch n.Tag {
		case "p":
			// Captions and linked summaries are not the sustained prose of an article.
			linked := false
			for p := n.Parent; p != nil && p != scope; p = p.Parent {
				if p.Tag == "a" {
					linked = true
					break
				}
			}
			if !linked && len(normalize(textOf(n))) >= 150 {
				prose++
			}
		case "a":
			label := normalize(textOf(n))
			if len(label) < 12 || len(label) > 240 {
				continue
			}
			href := strings.TrimSpace(n.Attr["href"])
			if href == "" || strings.HasPrefix(href, "#") {
				continue
			}
			parsed, err := url.Parse(href)
			if err != nil || (parsed.Scheme != "" && parsed.Scheme != "https" && parsed.Scheme != "http") {
				continue
			}
			path := strings.TrimRight(parsed.Path, "/")
			if path == "" {
				continue
			}
			links++
			targets[path] = true
			isHeadline := len(label) >= 25 && len(label) <= 180
			for _, child := range find(n, func(x *Node) bool { return x != n && (x.Tag == "h2" || x.Tag == "h3" || x.Tag == "h4") }) {
				if child != nil {
					isHeadline = true
					break
				}
			}
			for p := n.Parent; p != nil && p != scope; p = p.Parent {
				if p.Tag == "h2" || p.Tag == "h3" || p.Tag == "h4" {
					isHeadline = true
					break
				}
			}
			if isHeadline {
				headlineLinks++
			}
		}
	}
	// Distinct destinations, rather than repeated header/footer links, are the
	// critical indicator of a directory of independent stories.
	distinct := len(targets)
	if distinct < 12 || links < 18 || headlineLinks < 8 {
		return false
	}
	// A coherent article has sustained non-linked editorial prose. Its related
	// links must not reclassify it as a homepage.
	if prose >= 7 {
		return false
	}
	return (headlineLinks >= 12 && distinct >= 18 && links >= 24 && prose <= 5) ||
		(headlineLinks >= 8 && distinct >= 12 && links >= 18 && prose <= 2)
}
