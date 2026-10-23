package extraction

import (
 "clean-article-reader/internal/article"
  "clean-article-reader/internal/sourcehtml"
	"net/url"
 "clean-article-reader/internal/recovery"
	"strings"
)

// recoverEditorialContinuity replaces an incomplete Readability subtree only
// when source paragraphs independently corroborate its identity and order.
// It uses already downloaded HTML and does not broaden access or bypass gates.
func RecoverEditorialContinuity(raw string, selected []article.Block, base *url.URL) []article.Block {
	if base == nil || len(selected) == 0 {
		return selected
	}
	root := sourcehtml.Parse(raw)
	var retained []string
	for _, b := range selected {
		if b.Kind == "p" && len(sourcehtml.Normalize(b.Text)) >= 65 {
			retained = append(retained, sourcehtml.Normalize(b.Text))
		}
	}
	if len(retained) < 2 {
		return selected
	}
	var best []article.Block
	bestMatches := 0
	for _, scope := range sourcehtml.Find(root, func(n *sourcehtml.Node) bool {
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
		if matches > bestMatches || CountWords(candidate) > CountWords(best) {
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
			selectedText[sourcehtml.Normalize(b.Text)] = true
		}
	}
	if !recovery.ReplacementAllowed(selected, best, selectedText, selectedImages) {
		return recoverAnchoredEditorialTail(root, selected, base)
	}
	return best
}

func orderedEditorialMatches(retained []string, candidate []article.Block) int {
	return recovery.OrderedMatches(retained,candidate)
}

// editorialContinuityBlocks traverses one article element in source order.
// Advertisements are removed as individual subtrees, leaving adjacent editorial
// siblings intact. Gallery images receive their figure captions where available.
func editorialContinuityBlocks(scope *sourcehtml.Node, base *url.URL) []article.Block {
	var blocks []article.Block
	images := map[string]bool{}
	seenText := map[string]bool{}
	var walk func(*sourcehtml.Node)
	walk = func(n *sourcehtml.Node) {
		if n == nil {
			return
		}
		// Only prune containers that really are advertisements or site UI.
		// A generic gallery/wrapper must not suppress its editorial siblings.
		if sourcehtml.BlockedTags[n.Tag] || promotionalRegion(n) && !hasSubstantiveEditorialChildren(n) {
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
			b := SemanticTextBlock(n, base)
			if len(b.Text) < 25 || (IsAdvertisingText(b.Text) && len(strings.Fields(b.Text)) < 12) || seenText[sourcehtml.Normalize(b.Text)] {
				return
			}
			seenText[sourcehtml.Normalize(b.Text)] = true
			blocks = append(blocks, b)
			return
		case "img":
			// Ancestors with mixed editorial/ad content have already been
			// checked during traversal; reject only ad-only subtrees.
			link := ImageFromNode(n, base)
			if link == "" || images[link] {
				return
			}
			images[link] = true
			caption := sourcehtml.Normalize(n.Attr["alt"])
			for parent := n.Parent; parent != nil && parent != scope; parent = parent.Parent {
				if parent.Tag == "figure" {
					for _, child := range sourcehtml.Find(parent, func(x *sourcehtml.Node) bool { return x.Tag == "figcaption" }) {
						if label := sourcehtml.Normalize(sourcehtml.Text(child)); label != "" {
							caption = label
							break
						}
					}
					break
				}
			}
			blocks = append(blocks, article.Block{Kind: "img", URL: link, Text: caption})
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
func hasSubstantiveEditorialChildren(n *sourcehtml.Node) bool {
	if n == nil || n.Tag == "script" || n.Tag == "style" {
		return false
	}
	count := 0
	for _, p := range sourcehtml.Find(n, func(x *sourcehtml.Node) bool { return x.Tag == "p" }) {
		if len(sourcehtml.Normalize(sourcehtml.Text(p))) >= 75 {
			count++
			if count >= 2 {
				return true
			}
		}
	}
	return false
}

func insideFigure(n, stop *sourcehtml.Node) bool {
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
func recoverAnchoredEditorialTail(root *sourcehtml.Node, selected []article.Block, base *url.URL) []article.Block {
	if root == nil || base == nil {
		return selected
	}
	last := ""
	for _, b := range selected {
		if b.Kind == "p" && len(sourcehtml.Normalize(b.Text)) >= 65 {
			last = sourcehtml.Normalize(b.Text)
		}
	}
	if last == "" {
		return selected
	}
	var anchor *sourcehtml.Node
	for _, n := range sourcehtml.Find(root, func(n *sourcehtml.Node) bool { return n.Tag == "p" }) {
		t := sourcehtml.Normalize(sourcehtml.Text(n))
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
	var scope *sourcehtml.Node
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
			existing[sourcehtml.Normalize(b.Text)] = true
		}
	}
	started := false
	var tail []article.Block
	paragraphs := 0
	var walk func(*sourcehtml.Node)
	walk = func(n *sourcehtml.Node) {
		if n == nil {
			return
		}
		if n == anchor {
			started = true
			return
		}
		if sourcehtml.BlockedTags[n.Tag] || (sourcehtml.Noisy(n) && !hasSubstantiveEditorialChildren(n) && !hasEditorialChildOutsideNoise(n)) || (promotionalRegion(n) && !hasSubstantiveEditorialChildren(n) && !hasEditorialChildOutsideNoise(n)) {
			return
		}
		if n.Tag == "figcaption" {
			return
		}
		if n.Tag == "p" || n.Tag == "h2" || n.Tag == "h3" || n.Tag == "h4" || n.Tag == "blockquote" {
			if !started || insideFigure(n, scope) {
				return
			}
			b := SemanticTextBlock(n, base)
			k := sourcehtml.Normalize(b.Text)
			if len(k) < 25 || (IsAdvertisingText(k) && len(strings.Fields(k)) < 12) || existing[k] {
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
			link := ImageFromNode(n, base)
			if link == "" || existingImages[link] {
				return
			}
			existingImages[link] = true
			caption := sourcehtml.Normalize(n.Attr["alt"])
			for p := n.Parent; p != nil && p != scope; p = p.Parent {
				if p.Tag == "figure" {
					for _, c := range sourcehtml.Find(p, func(x *sourcehtml.Node) bool { return x.Tag == "figcaption" }) {
						if v := sourcehtml.Normalize(sourcehtml.Text(c)); v != "" {
							caption = v
							break
						}
					}
					break
				}
			}
			tail = append(tail, article.Block{Kind: "img", URL: link, Text: caption})
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
	return append(append([]article.Block{}, selected...), tail...)
}

// hasEditorialChildOutsideNoise allows editorial siblings in a noisy outer
// wrapper while keeping independently marked advertising subtrees excluded.
func hasEditorialChildOutsideNoise(scope *sourcehtml.Node) bool {
	for _, p := range sourcehtml.Find(scope, func(n *sourcehtml.Node) bool { return n.Tag == "p" }) {
		if len(sourcehtml.Normalize(sourcehtml.Text(p))) < 75 {
			continue
		}
		blocked := false
		for a := p.Parent; a != nil && a != scope; a = a.Parent {
			if sourcehtml.Noisy(a) || promotionalRegion(a) {
				blocked = true
				break
			}
		}
		if !blocked && !IsAdvertisingText(sourcehtml.Normalize(sourcehtml.Text(p))) {
			return true
		}
	}
	return false
}
