package assembly

import (
 "testing"

 "clean-article-reader/internal/article"
 "clean-article-reader/internal/extraction"
)

// The initial evidence and its nested inline slices must survive a complete
// assembly run even when an individual stage mutates existing block contents.
func TestRunDeepSliceIsolation(t *testing.T) {
 input := extraction.Result{
  Article: article.Article{Blocks: []article.Block{{
   Kind: "p", Text: "Source text",
   Parts: []article.TextPart{{Text: "source part", URL: "https://example.org/source"}},
   Links: []article.InlineLink{{Text: "source link", URL: "https://example.org/link"}},
  }}},
  Evidence: extraction.SourceEvidence{SourceURL: "https://example.org/article"},
 }
 stage := StageFunc{
  StageName: "modify-inline",
  Transform: func(a *article.Article, _ extraction.SourceEvidence, _ string) {
   a.Blocks[0].Text = "Reconciled"
   a.Blocks[0].Parts[0].Text = "revised part"
   a.Blocks[0].Links[0].URL = "https://example.org/updated"
  },
 }
 out := Run(input, "", stage)
 if input.Article.Blocks[0].Text != "Source text" ||
  input.Article.Blocks[0].Parts[0].Text != "source part" ||
  input.Article.Blocks[0].Links[0].URL != "https://example.org/link" {
  t.Fatalf("assembly mutated source extraction: %+v", input.Article.Blocks[0])
 }
 if out.Blocks[0].Text != "Reconciled" ||
  out.Blocks[0].Parts[0].Text != "revised part" ||
  out.Blocks[0].Links[0].URL != "https://example.org/updated" {
  t.Fatalf("stage output was not retained: %+v", out.Blocks[0])
 }
}
