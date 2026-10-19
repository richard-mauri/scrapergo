package main

import (
	"net/url"
	"regexp"
	"strings"
)

var linkMatchWords = regexp.MustCompile(`[\pL\pN]+`)

// linkTextKey compares publisher HTML text to the *already selected* article
// paragraph without depending on insignificant whitespace before punctuation.
func linkTextKey(s string) string {
	return strings.Join(linkMatchWords.FindAllString(strings.ToLower(s), -1), " ")
}

// recoverMissingArticleLinks restores hyperlinks when Readability's plain-text
// extraction or structured-document recovery has omitted link metadata.
// It never adds a new paragraph or a sidebar link: the full paragraph must
// already be present in the selected article blocks.
func recoverMissingArticleLinks(a *Article, raw, source string) {
	if a == nil {
		return
	}
	base, err := url.Parse(source)
	if err != nil {
		return
	}
	root := parseHTML(raw)
	for _, node := range find(root, func(n *Node) bool { return n.Tag == "p" || n.Tag == "li" }) {
		if inNoisy(node, nil) { continue }
  anchors := find(node, func(n *Node) bool { return n.Tag == "a" })
		if len(anchors) == 0 {
			continue
		}
		key := linkTextKey(textOf(node))
		if len(key) < 12 {
			continue
		}
		for i := range a.Blocks {
			block := &a.Blocks[i]
			if block.Kind == "img" || block.Kind == "topic-link" || linkTextKey(block.Text) != key {
				continue
			}
			// If the structured extractor already preserved links, leave it alone.
			hasURL := false
			for _, part := range block.Parts {
				if part.URL != "" {
					hasURL = true
					break
				}
			}
			if hasURL || len(block.Links) != 0 {
				break
			}
			seen := make(map[string]bool)
			for _, anchor := range anchors {
				label := normalize(textOf(anchor))
				href := resolveImageURL(anchor.Attr["href"], base)
				if label == "" || href == "" || seen[label+"|"+href] {
					continue
				}
				if !strings.Contains(block.Text, label) {
					continue
				}
				seen[label+"|"+href] = true
				block.Links = append(block.Links, InlineLink{Text: label, URL: href})
			}
			// The reader prioritizes Parts when present. Drop plain-only parts so
			// its established Links renderer can display the recovered anchors.
			if len(block.Links) > 0 {
				block.Parts = nil
			}
			break
		}
	}
}
