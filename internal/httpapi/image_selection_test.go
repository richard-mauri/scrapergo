package httpapi

import (
 "clean-article-reader/internal/article"

 "testing"

)

func TestArticleImageCandidatesUsesRenderedImages(t *testing.T) {
	a := article.Article{Blocks: []article.Block{{Kind: "p", URL: "https://example.com/no"}, {Kind: "img", URL: "https://example.com/inline.jpg"}, {Kind: "img", URL: "https://example.com/inline.jpg"}, {Kind: "img", URL: "https://example.com/second.png"}}}
	got := ImageCandidates(a)
	if len(got) != 2 || got[0] != "https://example.com/inline.jpg" || got[1] != "https://example.com/second.png" {
		t.Fatalf("candidates: %#v", got)
	}
	a.ImageURL = "https://example.com/hero.jpg"
	got = ImageCandidates(a)
	if len(got) != 3 || got[0] != a.ImageURL {
		t.Fatalf("hero priority: %#v", got)
	}
}
