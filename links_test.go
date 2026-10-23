package main

import (
 "clean-article-reader/internal/extraction"
 "clean-article-reader/internal/ui"
 "clean-article-reader/internal/links"
	"net/url"
	"strings"
	"testing"
)

func TestArticleInlineLinkRouting(t *testing.T) {
	base, _ := url.Parse("https://www.aljazeera.com/news/story")
	blocks := []Block{{Kind: "p", Text: "Read the California court filing for details."}}
	extraction.AttachInlineLinks(blocks, `<article><p>Read the <a href="/news/filing">California court filing</a> for details.</p><a href="javascript:alert(1)">unsafe</a></article>`, base)
	if len(blocks[0].Links) != 1 || blocks[0].Links[0].URL != "https://www.aljazeera.com/news/filing" {
		t.Fatalf("links: %+v", blocks[0].Links)
	}
	if !strings.Contains(ui.Template(), "function putLinked(") || !strings.Contains(ui.Template(), "encodeURIComponent(dest.href)") {
		t.Fatal("ScraperGo reader routing missing")
	}
}
func TestLinkableDocuments(t *testing.T) {
	if links.IsDocumentURL("https://example.com/report.pdf") || links.IsDocumentURL("javascript:alert(1)") {
		t.Fatal("unsafe document routing")
	}
	if !links.IsDocumentURL("https://example.com/news/related") {
		t.Fatal("article routing disabled")
	}
}
