package links_test

import (
 "clean-article-reader/internal/links"
 "testing"
)

func TestLinkableDocuments(t *testing.T) {
	if links.IsDocumentURL("https://example.com/report.pdf") || links.IsDocumentURL("javascript:alert(1)") {
		t.Fatal("unsafe document routing")
	}
	if !links.IsDocumentURL("https://example.com/news/related") {
		t.Fatal("article routing disabled")
	}
}
