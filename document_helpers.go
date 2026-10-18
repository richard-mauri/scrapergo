package main

import (
	"html"
	"net/url"
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
		link := resolveImageURL(picture, base)
		caption = normalize(caption)
		if link == "" || caption == "" {
			continue
		}
		result[caption] = link
	}
	return result
}

func semanticTextBlock(n *Node, base *url.URL) Block {
	var parts []TextPart
	var walk func(*Node, bool, bool, string)
	walk = func(x *Node, bold, italic bool, link string) {
		if blockedTags[x.Tag] {
			return
		}
		if x.Tag == "br" {
			parts = append(parts, TextPart{Text: "\n"})
			return
		}
		if x.Tag == "#text" {
			v := html.UnescapeString(x.Text)
			if v != "" {
				parts = append(parts, TextPart{Text: v, Bold: bold, Italic: italic, URL: link})
			}
			return
		}
		bold = bold || x.Tag == "strong" || x.Tag == "b"
		italic = italic || x.Tag == "em" || x.Tag == "i"
		if x.Tag == "a" {
			if u := resolveImageURL(x.Attr["href"], base); u != "" && linkableDocument(u) {
				link = u
			}
		}
		for _, child := range x.Kids {
			walk(child, bold, italic, link)
		}
	}
	walk(n, false, false, "")
	// Merge runs with the same formatting and normalize whitespace without
	// flattening newlines that separate independent story facts.
	var compact []TextPart
	for _, part := range parts {
		part.Text = strings.ReplaceAll(part.Text, "\r", "")
		part.Text = regexp.MustCompile(`[\t ]+`).ReplaceAllString(part.Text, " ")
		if part.Text == "" {
			continue
		}
		if len(compact) > 0 && compact[len(compact)-1].Bold == part.Bold && compact[len(compact)-1].Italic == part.Italic && compact[len(compact)-1].URL == part.URL {
			compact[len(compact)-1].Text += part.Text
		} else {
			compact = append(compact, part)
		}
	}
	var raw strings.Builder
	for _, part := range compact {
		raw.WriteString(part.Text)
	}
	text := strings.TrimSpace(raw.String())
	// Reconstitute runs against normalized outer whitespace. The rendering
	// path uses the original part stream, which retains authorial line breaks.
	return Block{Kind: n.Tag, Text: text, Parts: compact}
}

func nestedProse(n, scope *Node) bool {
	for p := n.Parent; p != nil && p != scope; p = p.Parent {
		if p.Tag == "p" || p.Tag == "li" || p.Tag == "blockquote" || p.Tag == "figcaption" {
			return true
		}
	}
	return false
}
