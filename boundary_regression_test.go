package main

import (
 articlemodel "clean-article-reader/internal/article"
 "clean-article-reader/internal/extraction"
	"fmt"
	"net/url"
	"strings"
	"testing"
)

func TestTrailingSiblingPromotionAndNearbyTopics(t *testing.T) {
	var s strings.Builder
	s.WriteString(`<main><article><p>Introductory explanation is a meaningful paragraph introducing this detailed article.</p>`)
	for i := 0; i < 8; i++ {
		fmt.Fprintf(&s, `<section><h2>Entry number %d</h2><p>Editorial prose for this entry provides important detail, analysis and evidence necessary for understanding.</p></section>`, i)
	}
	s.WriteString(`<p>Closing editorial discussion is the last content in this excellent article and should be retained.</p></article>`)
	s.WriteString(`<div class="related"><div><h3>In This Topic</h3></div><div class="chips">`)
	for i := 0; i < 8; i++ {
		fmt.Fprintf(&s, `<a href="/tags/%d"><img src="/logo%d.png">Tag %d</a>`, i, i, i)
	}
	s.WriteString(`</div></div></main>`)
	u, _ := url.Parse("https://example.com/story")
	blocks, ok := extraction.RecoverStructuredDocument(s.String(), []articlemodel.Block{{Kind: "p", Text: "only a partial intro"}}, u)
	if !ok {
		t.Fatal("no recover")
	}
	var topics int
	for _, b := range blocks {
		if b.Kind == "topic-link" {
			topics++
		}
	}
	if topics != 8 {
		t.Fatalf("expected 8 topic links got %d", topics)
	}
}

func TestTrimSiblingRecommendation(t *testing.T) {
	b := []articlemodel.Block{{Kind: "p", Text: "Legitimate last paragraph."}, {Kind: "img", Text: "College Football", URL: "https://example.com/icon.png"}, {Kind: "img", Text: "Next Story Image: Some unrelated report", URL: "https://example.com/teaser.jpg"}, {Kind: "h2", Text: "Some unrelated report"}}
	cleaned := extraction.TrimTrailingRecommendation(b)
	if len(cleaned) != 1 {
		t.Fatalf("expected editorial only, got %#v", cleaned)
	}
}
