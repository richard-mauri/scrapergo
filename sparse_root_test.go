package main

import (
 "testing"
 "clean-article-reader/internal/pageclassify"
)

func TestSparseRootLanding(t *testing.T) {
	raw := `<html><body><main><h1>Discover what is new</h1><img src="/hero.jpg"><a href="/shop">Shop</a><a href="/learn">Learn</a><a href="/about">About</a></main></body></html>`
	if !pageclassify.IsSparseRootLanding(raw, "https://example.test/", 105) {
		t.Fatal("missed sparse marketing homepage")
	}
	if pageclassify.IsSparseRootLanding(raw, "https://example.test/story", 105) {
		t.Fatal("must not treat a nested article URL as a root landing page")
	}
}
func TestSparseRootArticleProtected(t *testing.T) {
	article := `<html><main><article><h1>Long news report</h1><p>` + longPara + longPara + longPara + longPara + `</p><img src="/lead.jpg"><a href="/one">Reference</a><a href="/two">Source</a><a href="/three">Read more</a></article></main></html>`
	if pageclassify.IsSparseRootLanding(article, "https://example.test/", 150) {
		t.Fatal("misclassified substantive article")
	}
}
