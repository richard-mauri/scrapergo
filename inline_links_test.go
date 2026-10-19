package main

import (
	"strings"
	"testing"
)

func TestInlineLinksAndLists(t *testing.T) {
	raw := `<html><head><meta property="og:title" content="Example Act"></head><body><article><h2>Proposed legislation</h2><p>This bill provides essential support for households and families throughout the nation, including the communities that have endured severe economic hardship.</p><p>Lawmakers said the measure would provide health, housing, and public school funding to families. The discussion specifically included improvements to rural healthcare, hospitals, and public schools.</p><ul><li>Provide direct payments to eligible families</li><li>Expand health coverage for seniors</li></ul><p>Read the bill text <a href="/docs/bill.pdf">here</a>.</p><p>Read a <a href="https://example.com/docs/summary.pdf">summary</a> here.</p><p>Read the economic <a href="/docs/analysis.pdf">analysis</a> here.</p></article></body></html>`
	a, err := extract(raw, "https://example.com/press/release")
	if err != nil {
		t.Fatal(err)
	}
	links := map[string]bool{}
	list := []string{}
	for _, b := range a.Blocks {
		if b.Kind == "li" {
			list = append(list, b.Text)
		}
		for _, p := range b.Parts {
			if p.URL != "" {
				links[p.URL] = true
			}
		}
	}
	if len(list) != 2 {
		t.Fatalf("expected two distinct list items, got %v", list)
	}
	for _, want := range []string{"https://example.com/docs/bill.pdf", "https://example.com/docs/summary.pdf", "https://example.com/docs/analysis.pdf"} {
		if !links[want] {
			t.Errorf("missing link: %s", want)
		}
	}
	if !strings.Contains(a.Blocks[len(a.Blocks)-1].Text, "analysis") {
		t.Fatal("lost final paragraph")
	}
}
func TestRejectUnsafeInlineLinks(t *testing.T) {
	root := parseHTML(`<p><a href="javascript:alert(1)">Click</a> <a href="//example.org/doc.pdf">Document</a></p>`)
	ps := find(root, func(n *Node) bool { return n.Tag == "p" })
	parts := inlineTextParts(ps[0], "https://example.com/test")
	found := false
	for _, p := range parts {
		if strings.Contains(p.URL, "javascript:") {
			t.Fatal("unsafe URL emitted")
		}
		if p.URL == "https://example.org/doc.pdf" {
			found = true
		}
	}
	if !found {
		t.Fatal("protocol relative URL not resolved")
	}
}
