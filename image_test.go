package main

import (
 "clean-article-reader/internal/extraction"
 "clean-article-reader/internal/sourcehtml"
	"net/url"
	"strings"
	"testing"
)

func TestImageURLResolution(t *testing.T) {
	base, _ := url.Parse("https://example.com/a/b")
	if got := sourcehtml.ResolveURL("/large.jpg", base); got != "https://example.com/large.jpg" {
		t.Fatal(got)
	}
	if got := sourcehtml.ResolveURL("javascript:alert(1)", base); got != "" {
		t.Fatal(got)
	}
}
func TestExtractImageAndCaption(t *testing.T) {
	u, _ := url.Parse("https://example.com/post")
	photos := extraction.ArticleImages(`<article><figure><img src="/photo.jpg" alt="fallback"><figcaption>Photographer credit</figcaption></figure><img src="javascript:1"></article>`, u)
	if len(photos) != 1 || photos[0].URL != "https://example.com/photo.jpg" || photos[0].Text != "Photographer credit" {
		t.Fatalf("%+v", photos)
	}
}
func TestReadabilityWithImage(t *testing.T) {
	raw := `<html><head><meta property="og:image" content="/hero.jpg"></head><body><article><h1>Study</h1><figure><img src="/inside.jpg" alt="Research image"><figcaption>Field study</figcaption></figure><p>` + strings.Repeat("This article describes work that is interesting and contains several findings. ", 30) + `</p></article></body></html>`
	a, e := extraction.ExtractWithReadability(raw, "https://example.com/story", extract, inlineTextParts)
	if e != nil {
		t.Fatal(e)
	}
	if a.ImageURL != "https://example.com/hero.jpg" {
		t.Fatalf("hero: %s", a.ImageURL)
	}
}
