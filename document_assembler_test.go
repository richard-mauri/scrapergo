package main

import (
 articlemodel "clean-article-reader/internal/article"
 "clean-article-reader/internal/documentmedia"
 "clean-article-reader/internal/extraction"
 "clean-article-reader/internal/recovery"
 "clean-article-reader/internal/reader"
 "clean-article-reader/internal/structured"
 "encoding/json"
 "clean-article-reader/internal/media"
 "clean-article-reader/internal/contributors"
 "testing"

 extractionmodel "clean-article-reader/internal/extraction"
)

// This characterization test freezes the existing mutation order, so future
// stage moves cannot silently alter the resulting article document.
func TestAssembleArticleDocumentMatchesExistingPipeline(t *testing.T) {
 sourceURL := "https://example.org/story"
 raw := `<html><head><meta property="og:image" content="/meta.jpg"></head><body><main><article><h1>Story</h1><figure><picture><img src="/lead.jpg"></picture></figure><p>A complete first paragraph of ordinary article content appears here with contextual information and enough text for source processing.</p></article></main></body></html>`
 input := extractionmodel.Result{
  articlemodel.Article: articlemodel.Article{Title:"Story",URL:sourceURL,Byline:"Reporter",WordCount:21,Blocks:[]articlemodel.Block{{Kind:"p",Text:"A complete first paragraph of ordinary article content appears here with contextual information and enough text for source processing."}}},
  Evidence: extractionmodel.SourceEvidence{SourceURL:sourceURL,MetadataImageURL:"https://example.org/meta.jpg"},
 }
 expected := input.articlemodel.Article
 extraction.RecoverArticlePhotographs(&expected,raw,sourceURL)
 media.Deduplicate(&expected)
 recovery.RecoverMissingArticleLinks(&expected,raw,sourceURL)
 contributors.RecoverAuthorPortrait(&expected,raw,sourceURL)
 structured.RemoveSerializedPlayerBlocks(&expected)
 contributors.ClassifyContributorImages(&expected)
 recovery.ClassifyImagesFromSource(&expected,raw,sourceURL)
 contributors.ReconcileContributorByline(&expected)
 contributors.SeparateHeader(&expected)
 recovery.RecoverOpeningSourceFigure(&expected,raw,sourceURL)
 documentmedia.NormalizeDocument(&expected,media.SourceIdentity,input.Evidence.MetadataImageURL)

 got := reader.DefaultProcessor().Process(input,raw)
 wantJSON,_ := json.Marshal(expected)
 gotJSON,_ := json.Marshal(got)
 if string(wantJSON)!=string(gotJSON) {t.Fatalf("assembly changed document:\nwant %s\ngot  %s",wantJSON,gotJSON)}
 if got.ImageURL!="https://example.org/lead.jpg" {t.Fatalf("unexpected selected lead: %q",got.ImageURL)}
}
