package main

import (
	"fmt"
	"testing"
)

func TestPageIndexRecognition(t *testing.T) {
	html := `<html><main><h1>News homepage</h1>`
	for i := 0; i < 12; i++ {
		html += fmt.Sprintf(`<section><h2><a href="/story-%d">A fresh report on the latest event</a></h2><a href="/more-%d">Read our coverage and analysis</a></section>`, i, i)
	}
	html += `</main></html>`
	if !pageIsIndex(html) {
		t.Fatal("linked headline index should not be presented as one article")
	}
}
func TestArticleWithManyLinksIsNotIndex(t *testing.T) {
	html := `<html><main><h1>Research roundup</h1>`
	for i := 0; i < 12; i++ {
		html += `<section><h2><a href="/study">Research results and analysis</a></h2><p>This long discussion explains the study and its implications with enough detail to be counted as substantive editorial prose rather than a short navigational teaser. It also supplies additional background and helpful analysis for the reader.</p><a href="/source">Read the research source material</a></section>`
	}
	html += `</main></html>`
	if pageIsIndex(html) {
		t.Fatal("must preserve long-form structured articles")
	}
}
func TestArticleMetadataWithSustainedProseIsNotIndex(t *testing.T) {
	html := `<html><head><script type="application/ld+json">{"@type":"NewsArticle"}</script></head><main>`
	for i := 0; i < 12; i++ {
		html += fmt.Sprintf(`<h2><a href="/chapter-%d">Chapter of a technical reference</a></h2><p>This chapter contains sustained commentary explaining technical subjects in detail, with original editorial prose intended for reading. It provides enough contextual information to distinguish the main content from unrelated navigation and index cards.</p><a href="/detail-%d">More technical explanation here</a>`, i, i)
	}
	html += `</main></html>`
	if pageIsIndex(html) {
		t.Fatal("structured article metadata should protect actual articles")
	}
}

func TestHomepageWithAnchorWrappedHeadingsAndArticleSchemaElsewhere(t *testing.T) {
	html := `<html><head><script type="application/ld+json">{"@graph":[{"@type":"WebPage"},{"@type":"NewsArticle"}]}</script></head><main>`
	for i := 0; i < 28; i++ {
		html += fmt.Sprintf(`<section><a href="/news/story-%d"><h2>A different independent report from across the world</h2></a><a href="/news/detail-%d">More coverage about the unfolding events</a></section>`, i, i)
	}
	html += `</main></html>`
	if !pageIsIndex(html) {
		t.Fatal("story cards with anchor-wrapped headings are an index even if embedded metadata mentions an article")
	}
}
func TestHomepageWithDirectHeadlineLinks(t *testing.T) {
	html := `<html><main>`
	for i := 0; i < 35; i++ {
		html += fmt.Sprintf(`<div><a href="/world/report-%d">Breaking developments in independent world report number %d</a></div>`, i, i)
	}
	html += `</main></html>`
	if !pageIsIndex(html) {
		t.Fatal("direct headline links should be recognized as a news index")
	}
}
func TestEditorialProseProtectsRichArticle(t *testing.T) {
	html := `<html><main><h1>Long-form report with resources</h1>`
	for i := 0; i < 10; i++ {
		html += `<p>This reporting provides continuous editorial explanation rather than collections of external story cards. The discussion has several substantial paragraphs with detailed analysis of the issue, its causes, and its possible effects over time.</p>`
	}
	for i := 0; i < 35; i++ {
		html += fmt.Sprintf(`<a href="/references/%d">A useful supporting research document with context</a>`, i)
	}
	html += `</main></html>`
	if pageIsIndex(html) {
		t.Fatal("article with many supporting links must not be classified as homepage")
	}
}
