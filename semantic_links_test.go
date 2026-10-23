package main

import (
 "clean-article-reader/internal/extraction"
	"net/url"
	"strings"
	"testing"
)

func TestPreserveSelectedHTMLLinksDuplicateLabels(t *testing.T) {
	html := `<article><p>Read the bill text <a href="/docs/bill.pdf">here</a>.</p><p>Read a summary <a href="/docs/summary.pdf">here</a>.</p><p>Read the economic analysis <a href="/docs/economics.pdf">here</a>.</p><p>Visit <a href="/a">first</a> and <a href="/b">second</a>.</p></article>`
	b := []Block{{Kind: "p", Text: "Read the bill text here."}, {Kind: "p", Text: "Read a summary here."}, {Kind: "p", Text: "Read the economic analysis here."}, {Kind: "p", Text: "Visit first and second."}}
	u, _ := url.Parse("https://example.com/article")
	extraction.PreserveSelectedHTMLLinks(b, html, u, inlineTextParts)
	for i, expect := range []string{"/docs/bill.pdf", "/docs/summary.pdf", "/docs/economics.pdf"} {
		n := 0
		for _, p := range b[i].Parts {
			if strings.HasSuffix(p.URL, expect) {
				n++
			}
		}
		if n != 1 {
			t.Errorf("paragraph %d: wanted %s exactly once; parts=%+v", i, expect, b[i].Parts)
		}
	}
	got := []string{}
	for _, p := range b[3].Parts {
		if p.URL != "" {
			got = append(got, p.URL)
		}
	}
	if len(got) != 2 || got[0] != "https://example.com/a" || got[1] != "https://example.com/b" {
		t.Fatalf("two ordered links: %v", got)
	}
}
func TestPreserveSelectedHTMLLinksNoNoiseOrUnselected(t *testing.T) {
	html := `<nav><p>Read a summary <a href="/fake">here</a>.</p></nav><article><p>Read a summary <a href="/real.pdf">here</a>.</p><p>Not an article <a href="/else">link</a>.</p><p>Unsafe <a href="javascript:alert(1)">script</a>.</p></article>`
	b := []Block{{Kind: "p", Text: "Read a summary here."}}
	u, _ := url.Parse("https://example.com/")
	extraction.PreserveSelectedHTMLLinks(b, html, u, inlineTextParts)
	found := []string{}
	for _, p := range b[0].Parts {
		if p.URL != "" {
			found = append(found, p.URL)
		}
	}
	if len(found) != 1 || found[0] != "https://example.com/real.pdf" {
		t.Fatalf("wrong recovered links: %v", found)
	}
}

// Exercise the same Readability entry point used by /api/read. This regression
// intentionally uses ordinary HTML rather than publisher-specific markup.
func TestReadabilityPreservesDocumentLinksEndToEnd(t *testing.T) {
	prose := "The proposal describes its funding mechanisms, projected allocations, implementation timetable, and the public oversight process. "
	long := strings.Repeat(prose, 4)
	raw := `<html><head><title>Public policy summary</title></head><body><article><h1>Public policy summary</h1><p>` + long + `</p><p>Read the full proposal <a href="/docs/proposal.pdf">online</a>.</p><p>Read the executive summary <a href="/docs/summary.pdf">online</a>.</p><p>Read the underlying analysis <a href="/docs/analysis.pdf">online</a>.</p><p>` + long + `</p></article></body></html>`
	article, err := extraction.ExtractWithReadability(raw, "https://example.org/news/policy",extract,inlineTextParts)
	if err != nil {
		t.Fatalf("Readability extraction failed: %v", err)
	}
	want := map[string]string{"Read the full proposal online.": "/docs/proposal.pdf", "Read the executive summary online.": "/docs/summary.pdf", "Read the underlying analysis online.": "/docs/analysis.pdf"}
	for _, b := range article.Blocks {
		suffix, ok := want[b.Text]
		if !ok {
			continue
		}
		found := false
		for _, part := range b.Parts {
			if strings.HasSuffix(part.URL, suffix) {
				found = true
			}
		}
		if !found {
			for _, link := range b.Links {
				if strings.HasSuffix(link.URL, suffix) {
					found = true
				}
			}
		}
		if !found {
			t.Errorf("link missing for %q, parts=%+v links=%+v", b.Text, b.Parts, b.Links)
		}
		delete(want, b.Text)
	}
	for missing := range want {
		t.Errorf("paragraph not preserved: %q", missing)
	}
}
