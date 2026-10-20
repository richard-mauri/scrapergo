package main

import (
	"net/url"
	"strings"
	"testing"
)

func TestRecoverLeadingParagraphsPreservesOrderAndLinks(t *testing.T) {
	raw := `<html><body><nav><p>Unrelated navigation summary with enough words to look like a real paragraph but must not be included in the actual story.</p></nav><main><article><h1>Truck review</h1>` +
		`<p>As I sat in the cab of the pickup truck and manually rolled down the windows, I closed my eyes and remembered my dad's old truck.</p>` +
		`<p>Last week, in a former Tesla store staffed by former Tesla employees, Slate opened its first showroom in Chicago's Fulton Market District.</p>` +
		`<p>Designed for pickup buyers, the <a href="/cars/truck">Slate Truck is a bare-bones electric pickup truck</a> that will go on sale with very few options.</p>` +
		`<p>Upon walking into the showroom on Monday morning, the first thing I noticed was a giant illuminated sign displaying the vehicle's price.</p>` +
		`<p>The showroom has been open since October, and the response from prospective customers was described as extremely enthusiastic by its manager.</p>` +
		`<p>We are really happy with the deposits we have received since opening, the showroom manager explained to the visiting reporter that morning.</p>` +
		`<p>I was hoping there would be at least two trucks on site, so I could compare their build quality and see the panel gaps on each one.</p>` +
		`<p>But there was only a single black prototype on display, and the driver door was left open so that visitors could check out its interior.</p>` +
		`<p>It was bigger than I expected—to my eye, roughly the size of a Hyundai Santa Cruz, and it was impressively sturdy when I pressed on it.</p>` +
		`<p>The next paragraph remains exactly where it was after the first retained paragraph and should not be duplicated by the recovery logic.</p></article></main></body></html>`
	base, _ := url.Parse("https://example.org/cars/review")
	original := []Block{{Kind: "p", Text: "It was bigger than I expected—to my eye, roughly the size of a Hyundai Santa Cruz, and it was impressively sturdy when I pressed on it."}, {Kind: "p", Text: "The next paragraph remains exactly where it was after the first retained paragraph and should not be duplicated by the recovery logic."}}
	got := recoverLeadingParagraphs(raw, original, base)
	if len(got) != 10 {
		t.Fatalf("got %d paragraphs, expected 10", len(got))
	}
	if !strings.HasPrefix(got[0].Text, "As I sat in the cab") {
		t.Fatalf("wrong start: %q", got[0].Text)
	}
	if !strings.HasPrefix(got[8].Text, "It was bigger than I expected") {
		t.Fatalf("wrong boundary: %q", got[8].Text)
	}
	linked := false
	for _, part := range got[2].Parts {
		if part.URL == "https://example.org/cars/truck" {
			linked = true
		}
	}
	if !linked {
		t.Fatal("introductory article link was lost")
	}
}

func TestRecoverLeadingParagraphsDeclinesWithoutEditorialScope(t *testing.T) {
	base, _ := url.Parse("https://example.org/story")
	raw := `<html><body><p>Earlier paragraph from an unrelated section with plenty of words and other material that must not be imported.</p><p>The original first paragraph is sufficiently long to serve as a trusted anchor within the source page itself.</p></body></html>`
	blocks := []Block{{Kind: "p", Text: "The original first paragraph is sufficiently long to serve as a trusted anchor within the source page itself."}}
	if got := recoverLeadingParagraphs(raw, blocks, base); len(got) != 1 {
		t.Fatalf("unexpected recovery: %+v", got)
	}
}
