package structured

import (
 "clean-article-reader/internal/sourcehtml"
 "clean-article-reader/internal/links"
 "net/url"
 "strings"
 "testing"
)

func TestSemanticTextBlockKeepsMarkup(t *testing.T) {
	base, _ := url.Parse("https://example.org/guide")
	root := sourcehtml.Parse(`<p><strong>First:</strong> read the guide<br><a href="/details"><em>Second:</em> follow the link</a></p>`)
	ps := sourcehtml.Find(root, func(n *sourcehtml.Node) bool { return n.Tag == "p" })
	if len(ps) != 1 {
		t.Fatal("paragraph missing")
	}
	b := TextBlock(ps[0], base, sourcehtml.ResolveURL, links.IsDocumentURL)
	hasBold, hasNewline, hasLink := false, false, false
	for _, p := range b.Parts {
		if p.Bold && strings.Contains(p.Text, "First:") {
			hasBold = true
		}
		if strings.Contains(p.Text, "\n") {
			hasNewline = true
		}
		if strings.Contains(p.URL, "/details") {
			hasLink = true
		}
	}
	if !hasBold || !hasNewline || !hasLink {
		t.Fatalf("semantic formatting lost: %#v", b.Parts)
	}
}
