package main

import (
 "clean-article-reader/internal/fetcher"
 "clean-article-reader/internal/extraction"
	"context"
	"strings"
	"testing"
)

func TestExtraction(t *testing.T) {
	raw := `<html><head><meta property="og:title" content="Texas factories"><meta name="author" content="A. Writer"></head><body><nav>Menus</nav><article><h1>Texas factories</h1><div class="advertisement"><p>Buy now at a great price and claim this limited offer immediately.</p></div><p>A semiconductor supply chain expansion is underway in Texas, where companies are evaluating sites for new manufacturing and assembly partnerships. This paragraph contains enough material to count as relevant article text.</p><h2>Why it matters</h2><p>Moving component suppliers closer to their customers can reduce transport delays and shorten production schedules. The long-term impact depends on costs, supplier commitments, and how quickly plants reach meaningful capacity.</p><div class="newsletter"><p>Enter your email for a free report and get updates every morning.</p></div></article><footer><p>Read our other stories</p></footer></body></html>`
	a, err := extraction.LegacyExtract(raw, "https://example.com/a")
	if err != nil {
		t.Fatal(err)
	}
	if a.Title != "Texas factories" || a.Byline != "A. Writer" {
		t.Fatalf("bad metadata: %+v", a)
	}
	if len(a.Blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %#v", a.Blocks)
	}
	for _, b := range a.Blocks {
		if strings.Contains(b.Text, "Buy now") || strings.Contains(b.Text, "Enter your email") {
			t.Fatal("advertising text retained")
		}
	}
}
func TestRejectLocal(t *testing.T) {
	for _, u := range []string{"file:///etc/passwd", "http://localhost/private", "http://127.0.0.1/", "http://10.0.0.1/"} {
		if _, err := fetcher.ValidatedURL(context.Background(), u); err == nil {
			t.Errorf("accepted %s", u)
		}
	}
}

func TestJSONLDFallback(t *testing.T) {
	raw := `<html><head><meta property="og:title" content="Structured article"><script type="application/ld+json">{"@context":"https://schema.org","@type":"NewsArticle","articleBody":"A manufacturing company is planning a new facility in Texas to support its customers, with several distinct suppliers and a lengthy implementation schedule. The project requires permits and financing before construction begins.\n\nThe supplier investment is expected to support technology manufacturers by shortening their logistics chain and reducing shipment delays. The organizations involved have not announced a final opening date or binding purchase commitments."}</script></head><body><div class="sponsored-wrapper"><p>Visible article body can be missed by filters.</p></div></body></html>`
	a, err := extraction.LegacyExtract(raw, "https://example.com/article")
	if err != nil {
		t.Fatal(err)
	}
	if a.WordCount < 60 {
		t.Fatalf("unexpected short article: %+v", a)
	}
}
