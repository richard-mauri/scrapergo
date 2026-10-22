package main

import (
 "clean-article-reader/internal/sourcehtml"
 "clean-article-reader/internal/links"
		"net/url"
 "clean-article-reader/internal/structured"
	"regexp"
	"strconv"
	"strings"
)

// Helpers for generic ordered HTML recovery. No site or ranking assumptions.

func countWords(blocks []Block) int {
	n := 0
	for _, b := range blocks {
		if b.Kind != "img" {
			n += len(strings.Fields(b.Text))
		}
	}
	return n
}

func nodeDepth(n *Node) int {
	d := 0
	for n != nil {
		d++
		n = n.Parent
	}
	return d
}

// serializedCaptionImages recovers the relationship between photograph and caption
// from the serialized embedded structured data. It deliberately does not download anything.
// Matches only static photo assets followed immediately by descriptive text.
var serializedPhotoRE = regexp.MustCompile(`"(https?://[^"\\]+\.(?:jpg|jpeg|png|webp)(?:\?[^"\\]*)?)","((?:[^"\\]|\\.){30,600})"`)

func serializedCaptionImages(raw string, base *url.URL) map[string]string {
	result := map[string]string{}
	for _, m := range serializedPhotoRE.FindAllStringSubmatch(raw, -1) {
		picture, err1 := strconv.Unquote(`"` + m[1] + `"`)
		caption, err2 := strconv.Unquote(`"` + m[2] + `"`)
		if err1 != nil || err2 != nil {
			continue
		}
		link := sourcehtml.ResolveURL(picture, base)
		caption = normalize(caption)
		if link == "" || caption == "" {
			continue
		}
		result[caption] = link
	}
	return result
}

// Preserve legacy entry point while structured conversion is package-owned.
func semanticTextBlock(n *Node, base *url.URL) Block {
 return structured.TextBlock(n,base,sourcehtml.ResolveURL,links.IsDocumentURL)
}

func nestedProse(n, scope *Node) bool {
	for p := n.Parent; p != nil && p != scope; p = p.Parent {
		if p.Tag == "p" || p.Tag == "li" || p.Tag == "blockquote" || p.Tag == "figcaption" {
			return true
		}
	}
	return false
}
