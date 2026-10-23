// Package reader owns ScraperGo's ordered article processing workflow.
// Transport and UI code must consume article.Article rather than importing
// extraction, media, or source HTML implementation details.
package reader

import (
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/assembly"
 "clean-article-reader/internal/contributors"
 "clean-article-reader/internal/extraction"
 "clean-article-reader/internal/media"
 "clean-article-reader/internal/structured"
)

// SourceTransform receives the captured source HTML and its resolved URL.
// It must not independently retrieve source content.
type SourceTransform func(*article.Article, string, string)

// Dependencies contains the remaining source-specific transformations while
// they are migrated out of the historical executable package in PR #43.
type Dependencies struct {
 RecoverPhotographs SourceTransform
 RecoverLinks SourceTransform
 RecoverAuthorPortrait SourceTransform
 ClassifyContributors func(*article.Article)
 ClassifySourceImages SourceTransform
 ReconcileByline func(*article.Article)
 RecoverOpeningFigure SourceTransform
 Normalize func(*article.Article, string)
 RestoreGalleryGroups SourceTransform
}

// Processor owns stage construction, sequencing, and document isolation.
type Processor struct { dependencies Dependencies }

func New(dependencies Dependencies) Processor { return Processor{dependencies: dependencies} }

// Stages exposes the deterministic processing order for contract tests.
func (p Processor) Stages() []assembly.Stage {
 d := p.dependencies
 source := func(name string, fn SourceTransform) assembly.Stage {
  return assembly.StageFunc{StageName: name, Transform: func(a *article.Article, e extraction.SourceEvidence, html string) {
   fn(a, html, e.SourceURL)
  }}
 }
 plain := func(name string, fn func(*article.Article)) assembly.Stage {
  return assembly.StageFunc{StageName: name, Transform: func(a *article.Article, _ extraction.SourceEvidence, _ string) { fn(a) }}
 }
 return []assembly.Stage{
  source("recover-photographs", d.RecoverPhotographs),
  plain("deduplicate-images", media.Deduplicate),
  source("recover-links", d.RecoverLinks),
  source("recover-author-portrait", d.RecoverAuthorPortrait),
  plain("remove-player-blocks", structured.RemoveSerializedPlayerBlocks),
  plain("classify-contributors", d.ClassifyContributors),
  source("classify-source-images", d.ClassifySourceImages),
  plain("exclude-poststory-author-portraits", media.RemovePostEditorialAuthorPortraits),
  plain("reconcile-byline", d.ReconcileByline),
  plain("separate-contributor-header", contributors.SeparateHeader),
  source("recover-opening-figure", d.RecoverOpeningFigure),
  assembly.StageFunc{StageName:"normalize", Transform:func(a *article.Article, e extraction.SourceEvidence, _ string) {
   d.Normalize(a,e.MetadataImageURL)
  }},
  source("restore-gallery-groups", d.RestoreGalleryGroups),
 }
}

func (p Processor) Process(result extraction.Result, sourceHTML string) article.Article {
 return assembly.Run(result,sourceHTML,p.Stages()...)
}
