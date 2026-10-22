// Package assembly provides a testable, order-preserving pipeline for article
// document transformations without depending on HTTP or HTML parsing.
package assembly

import (
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/extraction"
)

// Stage transforms an article using already-captured source evidence.
// Stages must not perform independent network retrieval.
type Stage interface {
 Name() string
 Apply(*article.Article, extraction.SourceEvidence, string)
}

// Run executes ordered stages against an article value. Each invocation owns
// its own top-level document and block slice, protecting the extraction result
// from in-place mutations performed by legacy stages.
func Run(result extraction.Result, sourceHTML string, stages ...Stage) article.Article {
 document := result.Article
 document.Blocks = append([]article.Block(nil), document.Blocks...)
 for i := range document.Blocks {
  document.Blocks[i].Parts = append([]article.TextPart(nil), document.Blocks[i].Parts...)
  document.Blocks[i].Links = append([]article.InlineLink(nil), document.Blocks[i].Links...)
 }
 for _, stage := range stages {
  stage.Apply(&document, result.Evidence, sourceHTML)
 }
 return document
}

// StageFunc adapts existing processing functions to the shared contract.
type StageFunc struct {
 StageName string
 Transform func(*article.Article, extraction.SourceEvidence, string)
}

func (s StageFunc) Name() string { return s.StageName }
func (s StageFunc) Apply(a *article.Article, e extraction.SourceEvidence, html string) {
 s.Transform(a,e,html)
}
