package main

import (
 "clean-article-reader/internal/pageclassify"
	"fmt"
	"testing"
)

func TestShortNewsIndexClassified(t *testing.T) {
	s := `<html><main><h1>Latest news briefing</h1>`
	for i := 0; i < 12; i++ {
		s += fmt.Sprintf(`<section><a href="/stories/%d"><h2>Independent top report about today</h2><img src="/thumb%d.jpg"></a></section>`, i, i)
	}
	s += `</main></html>`
	if !pageclassify.IsThinPortal(s, 267) {
		t.Fatal("short news index should be identified")
	}
}
func TestShortMarketingPageClassified(t *testing.T) {
	s := `<html><main><h1>Products built for everyone</h1>`
	for i := 0; i < 4; i++ {
		s += fmt.Sprintf(`<section><h2>Explore our latest products %d</h2><p>Discover better tools for your everyday life.</p><img src="/product%d.jpg"><a href="/products/%d">Learn about this product</a></section>`, i, i, i)
	}
	s += `</main></html>`
	if !pageclassify.IsThinPortal(s, 105) {
		t.Fatal("short marketing home should be identified")
	}
}
func TestShortArticleNotClassified(t *testing.T) {
	s := `<html><main><article><h1>An important local development</h1><p>` + longPara + `</p><p>` + longPara + `</p><p>` + longPara + `</p><img src="/lead.jpg"></article></main></html>`
	if pageclassify.IsThinPortal(s, 155) {
		t.Fatal("genuine short article rejected")
	}
}
func TestSearchPortalRequiresSearchEvidence(t *testing.T) {
	if !pageclassify.IsSearchPortal(`<main><form role="search"><input name="q" type="search"></form><nav>Links</nav></main>`) {
		t.Fatal("search portal missed")
	}
	if pageclassify.IsSearchPortal(`<main><article><p>` + longPara + longPara + longPara + longPara + `</p></article><form role="search"><input name="q" type="search"></form></main>`) {
		t.Fatal("article with search should not be classified")
	}
}

const longPara = "This detailed reporting explains the background and consequences of the decision, including relevant historical evidence, interviews, and the specific circumstances that led to the outcome."
