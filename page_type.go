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

// pageIsLanding recognizes promotional/product landing pages whose main content
// is a collection of short feature pitches and calls to action, not an article.
// It is intentionally independent of hostname, URL path, and publisher text.
func pageIsLanding(raw string) bool {
	root := parseHTML(raw)
	scope := root
	for _, n := range find(root, func(n *Node) bool { return n.Tag == "main" }) {
		if !inNoisy(n, nil) {
			scope = n
			break
		}
	}
	headings, shortPitches, continuousProse, actions, sections := 0, 0, 0, 0, 0
	for _, n := range find(scope, func(n *Node) bool {
		switch n.Tag {
		case "h1", "h2", "h3", "p", "a", "section":
			return true
		}
		return false
	}) {
		if inNoisy(n, scope) {
			continue
		}
		s := normalize(textOf(n))
		switch n.Tag {
		case "section":
			sections++
		case "h1", "h2", "h3":
			if len(s) >= 12 && len(s) <= 180 {
				headings++
			}
		case "p":
			if len(s) >= 190 {
				continuousProse++
			} else if len(s) >= 35 && len(s) <= 185 {
				shortPitches++
			}
		case "a":
			href := strings.TrimSpace(n.Attr["href"])
			if href == "" || len(s) < 3 || len(s) > 55 {
				continue
			}
			if strings.HasPrefix(href, "#") || strings.HasPrefix(href, "/") || strings.HasPrefix(href, "http") {
				// Short, standalone actions in prominent section/hero content.
				for p := n.Parent; p != nil && p != scope; p = p.Parent {
					if p.Tag == "section" || p.Tag == "main" || p.Tag == "div" {
						if len(normalize(textOf(p))) <= 350 {
							actions++
						}
						break
					}
				}
			}
		}
	}
	// Require multiple independent feature pitches and multiple action targets.
	// Substantive paragraphs protect narrative articles and documentation.
	return headings >= 4 && shortPitches >= 3 && actions >= 3 &&
		continuousProse <= 2 && (sections >= 2 || headings >= 7)
}

// pageIsSearchPortal requires strong structural evidence of a search interface.
// An extraction error on its own never implies homepage classification.
func pageIsSearchPortal(raw string) bool {
	root := parseHTML(raw)
	searchForms, searchInputs, articleBodies := 0, 0, 0
	for _, n := range find(root, func(n *Node) bool { return n.Tag == "form" || n.Tag == "input" || n.Tag == "article" }) {
		switch n.Tag {
		case "form":
			if strings.EqualFold(n.Attr["role"], "search") || strings.Contains(strings.ToLower(n.Attr["action"]), "search") {
				searchForms++
			}
		case "input":
			if strings.EqualFold(n.Attr["type"], "search") || strings.EqualFold(n.Attr["name"], "q") {
				searchInputs++
			}
		case "article":
			if len(strings.Fields(textOf(n))) >= 90 {
				articleBodies++
			}
		}
	}
	return articleBodies == 0 && (searchForms >= 1 && searchInputs >= 1 || searchInputs >= 2)
}

// pageIsThinPortal checks the original document only when Readability returns
// a very short result. It demands multiple independent navigation or promotional
// groups and protects substantive article paragraphs and technical documentation.
func pageIsThinPortal(raw string, extractedWords int) bool {
	if extractedWords < 1 || extractedWords > 320 {
		return false
	}
	root := parseHTML(raw)
	scope := root
	for _, n := range find(root, func(n *Node) bool { return n.Tag == "main" }) {
		if !inNoisy(n, nil) {
			scope = n
			break
		}
	}
	headings, briefs, deep, images, links, groups := 0, 0, 0, 0, 0, 0
	destinations := map[string]bool{}
	for _, n := range find(scope, func(n *Node) bool {
		switch n.Tag {
		case "h1", "h2", "h3", "p", "img", "a", "section":
			return true
		}
		return false
	}) {
		if inNoisy(n, scope) {
			continue
		}
		v := normalize(textOf(n))
		switch n.Tag {
		case "h1", "h2", "h3":
			if len(v) >= 8 && len(v) <= 140 {
				headings++
			}
		case "p":
			if len(v) >= 180 {
				deep++
			} else if len(v) >= 20 && len(v) <= 140 {
				briefs++
			}
		case "img":
			images++
		case "section":
			groups++
		case "a":
			h := strings.TrimSpace(n.Attr["href"])
			if h != "" && len(v) >= 8 && len(v) <= 170 {
				links++
				destinations[h] = true
			}
		}
	}
	if deep >= 3 {
		return false
	}
	// Link-driven feed or directory: many independent destinations but few
	// continuous paragraphs, even if markup does not use conventional headings.
	feed := len(destinations) >= 8 && links >= 10 && deep <= 1 && (headings >= 2 || images >= 4 || groups >= 3)
	// Compact product/marketing landing page: several pitches or feature cards,
	// not one article with an illustration and a short caption.
	marketing := headings >= 3 && images >= 2 && links >= 3 && len(destinations) >= 3 && deep <= 1 && (briefs >= 2 || groups >= 3)
	return feed || marketing
}

// pageIsSparseRootLanding is a conservative fallback for short root landing
// pages that do not expose enough link/card structure for other classifiers.
// Hostnames and publisher-specific strings are intentionally not considered.
func pageIsSparseRootLanding(raw, source string, extractedWords int) bool {
	if extractedWords < 1 || extractedWords > 180 {
		return false
	}
	u, err := url.Parse(source)
	if err != nil || (u.Path != "" && u.Path != "/" && u.Path != "/index.html") {
		return false
	}
	root := parseHTML(raw)
	// An actual article element containing substantive prose wins, even on /.
	for _, n := range find(root, func(n *Node) bool { return n.Tag == "article" }) {
		if len(strings.Fields(normalize(textOf(n)))) >= 85 {
			return false
		}
	}
	// Likewise protect a sustained narrative even without semantic <article>.
	prose := 0
	for _, n := range find(root, func(n *Node) bool { return n.Tag == "p" }) {
		if inNoisy(n, nil) {
			continue
		}
		if len(normalize(textOf(n))) >= 160 {
			prose++
		}
		if prose >= 2 {
			return false
		}
	}
	// Require real evidence that it is a non-article landing page: a large
	// image/hero, multiple navigation links, or feature headings.
	images, links, headings := 0, 0, 0
	for _, n := range find(root, func(n *Node) bool {
		return n.Tag == "img" || n.Tag == "a" || n.Tag == "h2" || n.Tag == "h3"
	}) {
		if inNoisy(n, nil) {
			continue
		}
		switch n.Tag {
		case "img":
			images++
		case "a":
			if n.Attr["href"] != "" {
				links++
			}
		case "h2", "h3":
			headings++
		}
	}
	return (images >= 1 && links >= 3) || (headings >= 2 && links >= 2)
}
