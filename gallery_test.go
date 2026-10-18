package main

import (
	"net/url"
	"testing"
)

func TestGalleryImageCaptionAndOrder(t *testing.T) {
	u, _ := url.Parse("https://www.aljazeera.com/gallery/story")
	b := []Block{{Kind: "p", Text: "Opening paragraph."}, {Kind: "p", Text: "Closing paragraph."}}
	html := `<article><p>Opening paragraph.</p><figure><img src="/media/rally.jpg" alt="Crowd"><figcaption>Demonstrators rally at the Lincoln Memorial. [Jose Luis Magana/AP]</figcaption></figure><p>Closing paragraph.</p></article>`
	got := positionArticleFigures(b, html, u)
	if len(got) != 3 || got[1].Kind != "img" || got[1].Text != "Demonstrators rally at the Lincoln Memorial. [Jose Luis Magana/AP]" || got[1].URL != "https://www.aljazeera.com/media/rally.jpg" {
		t.Fatalf("wrong gallery order: %+v", got)
	}
}
func TestGalleryRawFigureFallback(t *testing.T) {
	u, _ := url.Parse("https://www.aljazeera.com/gallery/story")
	raw := `<main><figure><img src="/media/memorial.jpg"><figcaption>Demonstrators rally at the Lincoln Memorial in Washington, DC.</figcaption></figure></main>`
	got := supplementGalleryFigures(nil, raw, u)
	if len(got) != 1 || got[0].Kind != "img" {
		t.Fatalf("gallery figure missing: %+v", got)
	}
}
