package main

import (
	"strings"
	"testing"
)

func TestRecoverArticlePhotographsFromSiblingWrappers(t *testing.T) {
	html := `<main><div class="article-content"><p>Opening context describing a football match and providing sufficient material for a reader to keep.</p></div><div class="full-width-image mg-t-15 article-content"><picture><source srcset="/real-photo.jpg 1200w"><img src="/real-photo.jpg" alt="Players on field"></picture></div><p>Photo by Ed McGregor/FOX Sports</p><p>Following paragraph with adequate context about the match and what happened in the closing minutes.</p></main><aside><img src="/league-logo.png"></aside>`
	a := Article{Blocks: []Block{{Kind: "p", Text: "Opening context describing a football match and providing sufficient material for a reader to keep."}, {Kind: "p", Text: "Photo by Ed McGregor/FOX Sports"}, {Kind: "p", Text: "Following paragraph with adequate context about the match and what happened in the closing minutes."}}}
	recoverArticlePhotographs(&a, html, "https://example.org/story")
	count := 0
	for _, b := range a.Blocks {
		if b.Kind == "img" {
			count++
			if !strings.HasSuffix(b.URL, "/real-photo.jpg") {
				t.Fatalf("unexpected article image %s", b.URL)
			}
		}
	}
	if count != 1 {
		t.Fatalf("wanted one photo, got %d", count)
	}
}
