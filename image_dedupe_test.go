package main

import "testing"

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
	deduplicateArticleImages(a)
	if len(a.Blocks) != 5 {
		t.Fatalf("expected 5 blocks including 2 distinct inline photos, got %#v", a.Blocks)
	}
	if a.Blocks[2].URL != "https://images.example.org/content/second.jpg" || a.Blocks[4].URL != "https://images.example.org/content/third.jpg" {
		t.Fatalf("incorrect image order: %#v", a.Blocks)
	}
	if a.Blocks[1].Text != "Credit for hero" || a.Blocks[3].Text != "Second photograph credit" {
		t.Fatalf("credits lost: %#v", a.Blocks)
	}
}
func TestDifferentImagesRemain(t *testing.T) {
	a := &Article{Blocks: []Block{{Kind: "img", URL: "https://img.example.com/a.jpg"}, {Kind: "img", URL: "https://img.example.com/b.jpg"}}}
	deduplicateArticleImages(a)
	if len(a.Blocks) != 2 {
		t.Fatal("distinct photographs were removed")
	}
}
