package main

import (
 "clean-article-reader/internal/sourcehtml"
	"net/url"
	"regexp"
	"strings"
)

// navigationCluster rejects compact collections of linked topic chips, icons,
// and other page furniture. It does not use publisher names or heading text.
// Prose-rich editorial lists with explanatory links are deliberately retained.
func navigationCluster(n *Node) bool {
	switch n.Tag {
	case "section", "div", "ul", "ol":
	default:
		return false
	}
	anchors := find(n, func(x *Node) bool { return x.Tag == "a" && strings.TrimSpace(x.Attr["href"]) != "" })
	if len(anchors) < 5 || len(anchors) > 250 {
		return false
	}
	short := 0
	for _, a := range anchors {
		label := normalize(textOf(a))
		if len(label) > 0 && len(label) <= 65 && len(strings.Fields(label)) <= 8 {
			short++
		}
	}
	if short*100 < len(anchors)*80 {
		return false
	}
	// Real article paragraphs and explanatory list items must not be discarded.
	for _, p := range find(n, func(x *Node) bool { return x.Tag == "p" || x.Tag == "li" || x.Tag == "blockquote" }) {
		words := len(strings.Fields(normalize(textOf(p))))
		linked := 0
		for _, a := range find(p, func(x *Node) bool { return x.Tag == "a" }) {
			linked += len(strings.Fields(normalize(textOf(a))))
		}
		if words >= 18 || words-linked >= 9 {
			return false
		}
	}
	// A string of linked chips/icons normally dominates its container. Keep
	// substantial standalone editorial prose even when followed by links.
	totalWords := len(strings.Fields(normalize(textOf(n))))
	return totalWords <= len(anchors)*12+12
}

// inNavigationCluster checks whether a heading or other child belongs to a
// navigation group before it is counted as editorial content.
func inNavigationCluster(n, scope *Node) bool {
	for p := n; p != nil && p != scope; p = p.Parent {
		if navigationCluster(p) {
			return true
		}
	}
	return false
}

// promotionalRegion rejects a self-contained recommendation teaser without
// truncating the surrounding article. Text and attributes are both signals.
var promoLabelRE = regexp.MustCompile(`(?i)\b(next story|up next|read next|you may also like|recommended stories|more stories)\b`)

func promotionalRegion(n *Node) bool {
	if n.Tag != "section" && n.Tag != "div" && n.Tag != "h2" && n.Tag != "h3" {
		return false
	}
	if promoLabelRE.MatchString(n.Attr["aria-label"] + " " + n.Attr["data-testid"] + " " + n.Attr["class"] + " " + n.Attr["id"]) {
		return true
	}
	if n.Tag == "h2" || n.Tag == "h3" {
		return promoLabelRE.MatchString(normalize(textOf(n)))
	}
	heads := find(n, func(x *Node) bool { return x.Tag == "h2" || x.Tag == "h3" || x.Tag == "h4" })
	for _, h := range heads {
		if promoLabelRE.MatchString(normalize(textOf(h))) && len(strings.Fields(normalize(textOf(n)))) < 70 {
			return true
		}
	}
	return false
}

// contextualLinkSection recognizes labeled topical indexes rather than
// treating all link-dense content as editorial or all of it as noise.
// It requires a semantic heading and at least five short links in a cluster.
var topicLabelRE = regexp.MustCompile(`(?i)\b(topics?|tags?|categories|subjects?|in this|related topics?)\b`)

func contextualLinkSection(n *Node, base *url.URL) (string, []Block, bool) {
	if n.Tag != "section" && n.Tag != "div" {
		return "", nil, false
	}
	var title string
	for _, h := range find(n, func(x *Node) bool { return x.Tag == "h2" || x.Tag == "h3" || x.Tag == "h4" || x.Tag == "h5" }) {
		candidate := normalize(textOf(h))
		if len(candidate) > 0 && len(candidate) <= 80 && topicLabelRE.MatchString(candidate) {
			title = candidate
			break
		}
	}
	if title == "" {
		return "", nil, false
	}
	var cluster *Node
	for _, c := range find(n, func(x *Node) bool { return x != n && navigationCluster(x) }) {
		// Choose a compact cluster, not a large parent containing the article.
		if cluster == nil || len(find(c, func(x *Node) bool { return x.Tag == "a" })) < len(find(cluster, func(x *Node) bool { return x.Tag == "a" })) {
			cluster = c
		}
	}
	if cluster == nil {
		return "", nil, false
	}
	links := []Block{}
	seen := map[string]bool{}
	for _, a := range find(cluster, func(x *Node) bool { return x.Tag == "a" }) {
		label := normalize(textOf(a))
		href := sourcehtml.ResolveURL(a.Attr["href"], base)
		if label == "" || len(label) > 70 || href == "" || seen[href] {
			continue
		}
		seen[href] = true
		links = append(links, Block{Kind: "topic-link", Text: label, URL: href})
		if len(links) >= 80 {
			break
		}
	}
	return title, links, len(links) >= 5
}

func promotionalAncestor(n, stop *Node) bool {
	for p := n; p != nil && p != stop; p = p.Parent {
		if promotionalRegion(p) {
			return true
		}
	}
	return false
}
func topicHeading(n *Node) bool {
	if !topicLabelRE.MatchString(normalize(textOf(n))) {
		return false
	}
	for p := n.Parent; p != nil; p = p.Parent {
		if _, _, ok := contextualLinkSection(p, &url.URL{Scheme: "https", Host: "example.org"}); ok {
			return true
		}
	}
	return false
}

// A next-story teaser often appears as three sibling blocks rather than a
// nicely labeled enclosing section. If a promotional image caption occurs at
// the tail of an article, remove the preceding linked icon and its teaser.
func trimTrailingRecommendation(blocks []Block) []Block {
	for i := len(blocks) - 1; i >= 0; i-- {
		if i < len(blocks)-12 {
			break
		}
		b := blocks[i]
		if !promoLabelRE.MatchString(b.Text) {
			continue
		}
		// The marker must identify a recommendation, not an editorial discussion.
		if b.Kind != "img" && !strings.Contains(strings.ToLower(b.Text), "image:") {
			continue
		}
		start := i
		if start > 0 && blocks[start-1].Kind == "img" {
			start--
		}
		if start > 0 && blocks[start-1].Kind == "img" {
			start--
		}
		return blocks[:start]
	}
	return blocks
}

func hasContextualIndex(blocks []Block) bool {
	for _, b := range blocks {
		if b.Kind == "topic-title" {
			return true
		}
	}
	return false
}

// Find topic directories adjacent to a recovered editorial subtree. Restrict
// this to containers near the article rather than the whole site's header.
func findNearbyContextualIndex(root, scope *Node, base *url.URL) (string, []Block, bool) {
	var parent *Node
	if scope != nil {
		parent = scope.Parent
	}
	for level := 0; level < 3 && parent != nil; level, parent = level+1, parent.Parent {
		for _, n := range find(parent, func(n *Node) bool { return n.Tag == "section" || n.Tag == "div" }) {
			if promotionalRegion(n) || promotionalAncestor(n, parent) {
				continue
			}
			title, links, ok := contextualLinkSection(n, base)
			if ok {
				return title, links, true
			}
		}
	}
	return "", nil, false
}
