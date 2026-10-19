package main

import (
	"strings"
	"testing"
)

func TestRecoverMissingArticleLinksFromSelectedParagraphs(t *testing.T) {
	raw := `<html><body><nav><p>Read the bill text <a href="/unrelated.pdf">here</a>.</p></nav><article><p>Read the bill text <a href="/docs/bill.pdf">here</a>.</p><p>Read a summary <a href="/docs/summary.pdf">here</a>.</p><p>Read the economic analysis <a href="/docs/analysis.pdf">here</a>.</p><p>Other article paragraphs and information remain exactly as they were originally extracted.</p></article></body></html>`
	a := Article{Blocks: []Block{{Kind: "p", Text: "Read the bill text here."}, {Kind: "p", Text: "Read a summary here."}, {Kind: "p", Text: "Read the economic analysis here."}, {Kind: "p", Text: "Other article paragraphs and information remain exactly as they were originally extracted."}}}
	recoverMissingArticleLinks(&a, raw, "https://example.com/releases/story")
	if len(a.Blocks) != 4 {
		t.Fatalf("unexpected article blocks %d", len(a.Blocks))
	}
	for i, want := range []string{"/docs/bill.pdf", "/docs/summary.pdf", "/docs/analysis.pdf"} {
		if len(a.Blocks[i].Links) != 1 || !strings.HasSuffix(a.Blocks[i].Links[0].URL, want) {
			t.Errorf("block %d links = %+v, want %s", i, a.Blocks[i].Links, want)
		}
	}
	if len(a.Blocks[3].Links) != 0 {
		t.Fatal("unrelated content acquired links")
	}
}
func TestRecoverLinksRejectUnsafeAndDoNotReplaceExisting(t *testing.T) {
	raw := `<article><p>Read the document <a href="javascript:alert(1)">here</a>.</p><p>Another document <a href="/safe.pdf">here</a>.</p></article>`
	a := Article{Blocks: []Block{{Kind: "p", Text: "Read the document here."}, {Kind: "p", Text: "Another document here.", Parts: []TextPart{{Text: "Another document "}, {Text: "here", URL: "https://example.com/original.pdf"}, {Text: "."}}}}}
	recoverMissingArticleLinks(&a, raw, "https://example.com/article")
	if len(a.Blocks[0].Links) != 0 {
		t.Fatal("javascript link recovered")
	}
	if a.Blocks[1].Parts[1].URL != "https://example.com/original.pdf" {
		t.Fatal("existing link overwritten")
	}
}
