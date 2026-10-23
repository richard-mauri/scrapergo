package main

import (
 articlemodel "clean-article-reader/internal/article"
 "clean-article-reader/internal/bootstrap"
 "strings"
 "testing"
)

func TestExtractWithEvidencePreservesMetadataSeparately(t *testing.T) {
 prose := strings.Repeat("A carefully reported story includes clear details about the subject and presents reliable factual context for the reader. ", 13)
 raw := `<!doctype html><html><head><title>Example news report</title><meta property="og:type" content="article"><meta property="og:image" content="/lead.jpg"></head><body><main><article><h1>Example news report</h1><p>` + prose + `</p></article></main></body></html>`
 result, err := bootstrap.ExtractWithEvidence(raw, "https://example.org/news/example")
 if err != nil { t.Fatalf("extract: %v", err) }
 if result.Evidence.SourceURL != "https://example.org/news/example" { t.Fatalf("source URL: %q", result.Evidence.SourceURL) }
 if result.Evidence.MetadataImageURL != "https://example.org/lead.jpg" { t.Fatalf("metadata image: %q", result.Evidence.MetadataImageURL) }
 if result.Article.WordCount < 60 { t.Fatalf("missing article text: %+v", result.Article) }
 if result.Article.Blocks == nil { t.Fatal("expected article blocks") }
}
