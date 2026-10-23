package main

import (
 "clean-article-reader/internal/extraction"
 "clean-article-reader/internal/sourcehtml"
	"net/url"
 "clean-article-reader/internal/recovery"
	"strings"
)

// recoverEditorialContinuity replaces an incomplete Readability subtree only
// when source paragraphs independently corroborate its identity and order.
// It uses already downloaded HTML and does not broaden access or bypass gates.
func recoverEditorialContinuity(raw string, selected []Block, base *url.URL) []Block {
	if base == nil || len(selected) == 0 {
		return selected
	}
	root := parseHTML(raw)
	var retained []string
	for _, b := range selected {
		if b.Kind == "p" && len(normalize(b.Text)) >= 65 {
			retained = append(retained, normalize(b.Text))
		}
	}
	if len(retained) < 2 {
		return selected
	}
	var best []Block
	bestMatches := 0
	for _, scope := range find(root, func(n *Node) bool {
		return n.Tag == "article" || n.Tag == "main" || (n.Tag == "div" && strings.Contains(strings.ToLower(n.Attr["class"]+" "+n.Attr["id"]), "article-body"))
	}) {
		if sourcehtml.InNoisy(scope, nil) {
			continue
		}
		candidate := editorialContinuityBlocks(scope, base)
		matches := orderedEditorialMatches(retained, candidate)
		// Multiple independent anchors must agree and cover a substantial
		// fraction of the Readability paragraphs, in the correct source order.
		if matches < 2 || matches*2 < len(retained) || matches < bestMatches {
			continue
		}
		if matches > bestMatches || countWords(candidate) > countWords(best) {
			best, bestMatches = candidate, matches
		}
	}
	if bestMatches < 2 {
		return recoverAnchoredEditorialTail(root, selected, base)
	}
	selectedText := map[string]bool{}
	selectedImages := map[string]bool{}
	for _, b := range selected {
		if b.Kind == "img" {
			selectedImages[b.URL] = true
		} else if b.Kind == "p" {
			selectedText[normalize(b.Text)] = true
		}
	}
	if !recovery.ReplacementAllowed(selected, best, selectedText, selectedImages) {
		return recoverAnchoredEditorialTail(root, selected, base)
	}
	return best
}

func orderedEditorialMatches(retained []string, candidate []Block) int {
	return recovery.OrderedMatches(retained,candidate)
}

// editorialContinuityBlocks traverses one article element in source order.
// Advertisements are removed as individual subtrees, leaving adjacent editorial
// siblings intact. Gallery images receive their figure captions where available.
func editorialContinuityBlocks(scope *Node, base *url.URL) []Block {
	var blocks []Block
	images := map[string]bool{}
	seenText := map[string]bool{}
	var walk func(*Node)
	walk = func(n *Node) {
		if n == nil {
			return
		}
		// Only prune containers that really are advertisements or site UI.
		// A generic gallery/wrapper must not suppress its editorial siblings.
		if blockedTags[n.Tag] || promotionalRegion(n) && !hasSubstantiveEditorialChildren(n) {
			return
		}
		if sourcehtml.Noisy(n) && !hasSubstantiveEditorialChildren(n) {
			return
		}
		switch n.Tag {
		case "script", "style", "noscript", "iframe", "svg", "nav", "aside", "footer", "header", "form", "button":
			return
		case "p", "h2", "h3", "h4", "blockquote", "li":
			if nestedProse(n, scope) || insideFigure(n, scope) {
				return
			}
			b := semanticTextBlock(n, base)
			if len(b.Text) < 25 || (extraction.IsAdvertisingText(b.Text) && len(strings.Fields(b.Text)) < 12) || seenText[normalize(b.Text)] {
				return
			}
			seenText[normalize(b.Text)] = true
			blocks = append(blocks, b)
			return
		case "img":
			// Ancestors with mixed editorial/ad content have already been
			// checked during traversal; reject only ad-only subtrees.
			link := extraction.ImageFromNode(n, base)
			if link == "" || images[link] {
				return
			}
			images[link] = true
			caption := normalize(n.Attr["alt"])
			for parent := n.Parent; parent != nil && parent != scope; parent = parent.Parent {
				if parent.Tag == "figure" {
					for _, child := range find(parent, func(x *Node) bool { return x.Tag == "figcaption" }) {
						if label := normalize(textOf(child)); label != "" {
							caption = label
							break
						}
					}
					break
				}
			}
			blocks = append(blocks, Block{Kind: "img", URL: link, Text: caption})
			return
		case "figcaption":
			// Figure captions are attached to image blocks once, not emitted
			// again as prose (or separately as photographer credits).
			return
		}
		for _, child := range n.Kids {
			walk(child)
		}
	}
	walk(scope)
	return blocks
}

// Distinguish an ad-only wrapper from a mixed editorial wrapper. This is
// deliberately bounded to avoid deep scans on large publisher pages.
func hasSubstantiveEditorialChildren(n *Node) bool {
	if n == nil || n.Tag == "script" || n.Tag == "style" {
		return false
	}
	count := 0
	for _, p := range find(n, func(x *Node) bool { return x.Tag == "p" }) {
		if len(normalize(textOf(p))) >= 75 {
			count++
			if count >= 2 {
				return true
			}
		}
	}
	return false
}

func insideFigure(n, stop *Node) bool {
	for p := n.Parent; p != nil && p != stop; p = p.Parent {
		if p.Tag == "figure" || p.Tag == "figcaption" {
			return true
		}
	}
	return false
}

// recoverAnchoredEditorialTail preserves the selected introduction while
// recovering trailing siblings when the last selected paragraph is an exact
// anchor in the original HTML. No speculative expansion without an article
// scope and at least two substantial editorial paragraphs after the anchor.
func recoverAnchoredEditorialTail(root *Node, selected []Block, base *url.URL) []Block {
	if root == nil || base == nil {
		return selected
	}
	last := ""
	for _, b := range selected {
		if b.Kind == "p" && len(normalize(b.Text)) >= 65 {
			last = normalize(b.Text)
		}
	}
	if last == "" {
		return selected
	}
	var anchor *Node
	for _, n := range find(root, func(n *Node) bool { return n.Tag == "p" }) {
		t := normalize(textOf(n))
		if t == last || (len(t) >= 65 && len(last) >= 65 && (strings.HasPrefix(t, last) || strings.HasPrefix(last, t))) {
			if anchor != nil {
				return selected
			} // Ambiguous anchor.
			anchor = n
		}
	}
	if anchor == nil {
		return selected
	}
	var scope *Node
	for p := anchor.Parent; p != nil; p = p.Parent {
		if p.Tag == "article" {
			scope = p
			break
		}
		if scope == nil && (p.Tag == "main" || strings.Contains(strings.ToLower(p.Attr["class"]+" "+p.Attr["id"]), "article-body")) {
			scope = p
		}
	}
	if scope == nil {
		return selected
	}
	existing := map[string]bool{}
	existingImages := map[string]bool{}
	for _, b := range selected {
		if b.Kind == "img" {
			existingImages[b.URL] = true
		} else {
			existing[normalize(b.Text)] = true
		}
	}
	started := false
	var tail []Block
	paragraphs := 0
	var walk func(*Node)
	walk = func(n *Node) {
		if n == nil {
			return
		}
		if n == anchor {
			started = true
			return
		}
		if blockedTags[n.Tag] || (sourcehtml.Noisy(n) && !hasSubstantiveEditorialChildren(n) && !hasEditorialChildOutsideNoise(n)) || (promotionalRegion(n) && !hasSubstantiveEditorialChildren(n) && !hasEditorialChildOutsideNoise(n)) {
			return
		}
		if n.Tag == "figcaption" {
			return
		}
		if n.Tag == "p" || n.Tag == "h2" || n.Tag == "h3" || n.Tag == "h4" || n.Tag == "blockquote" {
			if !started || insideFigure(n, scope) {
				return
			}
			b := semanticTextBlock(n, base)
			k := normalize(b.Text)
			if len(k) < 25 || (extraction.IsAdvertisingText(k) && len(strings.Fields(k)) < 12) || existing[k] {
				return
			}
			existing[k] = true
			tail = append(tail, b)
			if b.Kind == "p" && len(k) >= 65 {
				paragraphs++
			}
			return
		}
		if n.Tag == "img" {
			if !started {
				return
			}
			link := extraction.ImageFromNode(n, base)
			if link == "" || existingImages[link] {
				return
			}
			existingImages[link] = true
			caption := normalize(n.Attr["alt"])
			for p := n.Parent; p != nil && p != scope; p = p.Parent {
				if p.Tag == "figure" {
					for _, c := range find(p, func(x *Node) bool { return x.Tag == "figcaption" }) {
						if v := normalize(textOf(c)); v != "" {
							caption = v
							break
						}
					}
					break
				}
			}
			tail = append(tail, Block{Kind: "img", URL: link, Text: caption})
			return
		}
		for _, child := range n.Kids {
			walk(child)
		}
	}
	walk(scope)
	if !started || paragraphs < 2 || len(tail) > 80 {
		return selected
	}
	return append(append([]Block{}, selected...), tail...)
}

// hasEditorialChildOutsideNoise allows editorial siblings in a noisy outer
// wrapper while keeping independently marked advertising subtrees excluded.
func hasEditorialChildOutsideNoise(scope *Node) bool {
	for _, p := range find(scope, func(n *Node) bool { return n.Tag == "p" }) {
		if len(normalize(textOf(p))) < 75 {
			continue
		}
		blocked := false
		for a := p.Parent; a != nil && a != scope; a = a.Parent {
			if sourcehtml.Noisy(a) || promotionalRegion(a) {
				blocked = true
				break
			}
		}
		if !blocked && !extraction.IsAdvertisingText(normalize(textOf(p))) {
			return true
		}
	}
	return false
}
