package main

import "testing"

func TestArticleImageCandidatesUsesRenderedImages(t *testing.T) {
	a := Article{Blocks: []Block{{Kind: "p", URL: "https://example.com/no"}, {Kind: "img", URL: "https://example.com/inline.jpg"}, {Kind: "img", URL: "https://example.com/inline.jpg"}, {Kind: "img", URL: "https://example.com/second.png"}}}
	got := articleImageCandidates(a)
	if len(got) != 2 || got[0] != "https://example.com/inline.jpg" || got[1] != "https://example.com/second.png" {
		t.Fatalf("candidates: %#v", got)
	}
	a.ImageURL = "https://example.com/hero.jpg"
	got = articleImageCandidates(a)
	if len(got) != 3 || got[0] != a.ImageURL {
		t.Fatalf("hero priority: %#v", got)
	}
}
