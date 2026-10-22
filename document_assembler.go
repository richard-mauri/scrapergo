package main

import (
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/assembly"
 "clean-article-reader/internal/extraction"
 "clean-article-reader/internal/media"
)

// assembleArticleDocument coordinates document transformations under a shared
// stage contract. The v0.2.85 processing order remains unchanged.
func assembleArticleDocument(extracted extraction.Result, sourceHTML string) Article {
 sourceStage := func(name string, fn func(*Article, string, string)) assembly.Stage {
  return assembly.StageFunc{StageName:name, Transform:func(a *article.Article,e extraction.SourceEvidence,html string){fn(a,html,e.SourceURL)}}
 }
 plainStage := func(name string,fn func(*Article)) assembly.Stage {
  return assembly.StageFunc{StageName:name,Transform:func(a *article.Article,_ extraction.SourceEvidence,_ string){fn(a)}}
 }
 stages := []assembly.Stage{
  sourceStage("recover-photographs",recoverArticlePhotographs),
  plainStage("deduplicate-images",media.Deduplicate),
  sourceStage("recover-links",recoverMissingArticleLinks),
  sourceStage("recover-author-portrait",recoverAuthorPortrait),
  plainStage("remove-player-blocks",removeSerializedPlayerBlocks),
  plainStage("classify-contributors",classifyContributorImages),
  sourceStage("classify-source-images",classifyImagesFromSource),
  plainStage("reconcile-byline",reconcileContributorByline),
  plainStage("separate-contributor-header",separateContributorHeader),
  sourceStage("recover-opening-figure",recoverOpeningSourceFigure),
  assembly.StageFunc{StageName:"normalize",Transform:func(a *article.Article,e extraction.SourceEvidence,_ string){normalizeArticleDocument(a,e.MetadataImageURL)}},
 }
 return assembly.Run(extracted,sourceHTML,stages...)
}
