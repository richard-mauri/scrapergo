// Package structured owns conversions from parsed source elements to article blocks.
package structured

import (
 "html"
 "net/url"
 "regexp"
 "strings"

 "clean-article-reader/internal/article"
 "clean-article-reader/internal/sourcehtml"
)

// TextBlock preserves structured inline parts while delegating URL policy to
// the caller. No HTML fetching or publisher-specific rules are performed.
func TextBlock(n *sourcehtml.Node, base *url.URL, resolve func(string, *url.URL) string, linkable func(string) bool) article.Block {
	var parts []article.TextPart
	var walk func(*sourcehtml.Node, bool, bool, string)
	walk = func(x *sourcehtml.Node, bold, italic bool, link string) {
		if sourcehtml.BlockedTags[x.Tag] {
			return
		}
		if x.Tag == "br" {
			parts = append(parts, article.TextPart{Text: "\n"})
			return
		}
		if x.Tag == "#text" {
			v := html.UnescapeString(x.Text)
			if v != "" {
				parts = append(parts, article.TextPart{Text: v, Bold: bold, Italic: italic, URL: link})
			}
			return
		}
		bold = bold || x.Tag == "strong" || x.Tag == "b"
		italic = italic || x.Tag == "em" || x.Tag == "i"
		if x.Tag == "a" {
			if u := resolve(x.Attr["href"], base); u != "" && linkable(u) {
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
	var compact []article.TextPart
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
	return article.Block{Kind: n.Tag, Text: text, Parts: compact}
}

