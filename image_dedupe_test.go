package main

import ("testing"
 "clean-article-reader/internal/media"
)

func TestDeduplicateHeroAndInlineImages(t *testing.T) {
	a := &Article{ImageURL: "https://images.example.org/content/photo-1200x800.jpg?width=1200", Blocks: []Block{
		{Kind: "p", Text: "Opening"},
		{Kind: "img", URL: "https://images.example.org/content/photo.jpg?width=600"},
		{Kind: "p", Text: "Credit for hero"},
		{Kind: "img", URL: "https://images.example.org/content/second.jpg"},
		{Kind: "img", URL: "https://images.example.org/content/second.jpg?size=small"},
		{Kind: "p", Text: "Second photograph credit"},
		{Kind: "img", URL: "https://images.example.org/content/third.jpg"},
	}}
	media.Deduplicate(a)
	if len(a.Blocks) != 6 {
		t.Fatalf("expected 6 blocks including metadata-matching lead and 2 distinct inline photos, got %#v", a.Blocks)
	}
	if a.Blocks[1].URL != "https://images.example.org/content/photo.jpg?width=600" || a.Blocks[3].URL != "https://images.example.org/content/second.jpg" || a.Blocks[5].URL != "https://images.example.org/content/third.jpg" {
		t.Fatalf("incorrect image order: %#v", a.Blocks)
	}
	if a.Blocks[2].Text != "Credit for hero" || a.Blocks[4].Text != "Second photograph credit" {
		t.Fatalf("credits lost: %#v", a.Blocks)
	}
}
func TestDifferentImagesRemain(t *testing.T) {
	a := &Article{Blocks: []Block{{Kind: "img", URL: "https://img.example.com/a.jpg"}, {Kind: "img", URL: "https://img.example.com/b.jpg"}}}
	media.Deduplicate(a)
	if len(a.Blocks) != 2 {
		t.Fatal("distinct photographs were removed")
	}
}
