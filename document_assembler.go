package main

import (
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/assembly"
 "clean-article-reader/internal/structured"
 "clean-article-reader/internal/extraction"
 "clean-article-reader/internal/contributors"
 "clean-article-reader/internal/media"
)

// assembleArticleDocument coordinates document transformations under a shared
// stage contract. The v0.2.85 processing order remains unchanged.
func articleAssemblyStages() []assembly.Stage {
 sourceStage := func(name string, fn func(*Article, string, string)) assembly.Stage {
  return assembly.StageFunc{StageName:name, Transform:func(a *article.Article,e extraction.SourceEvidence,html string){fn(a,html,e.SourceURL)}}
 }
 plainStage := func(name string,fn func(*Article)) assembly.Stage {
  return assembly.StageFunc{StageName:name,Transform:func(a *article.Article,_ extraction.SourceEvidence,_ string){fn(a)}}
 }
 return []assembly.Stage{
  sourceStage("recover-photographs",recoverArticlePhotographs),
  plainStage("deduplicate-images",media.Deduplicate),
  sourceStage("recover-links",recoverMissingArticleLinks),
  sourceStage("recover-author-portrait",recoverAuthorPortrait),
  plainStage("remove-player-blocks",structured.RemoveSerializedPlayerBlocks),
  plainStage("classify-contributors",classifyContributorImages),
  sourceStage("classify-source-images",classifyImagesFromSource),
  plainStage("exclude-poststory-author-portraits",media.RemovePostEditorialAuthorPortraits),
  plainStage("reconcile-byline",reconcileContributorByline),
  plainStage("separate-contributor-header",contributors.SeparateHeader),
  sourceStage("recover-opening-figure",recoverOpeningSourceFigure),
  assembly.StageFunc{StageName:"normalize",Transform:func(a *article.Article,e extraction.SourceEvidence,_ string){normalizeArticleDocument(a,e.MetadataImageURL)}},
  sourceStage("restore-gallery-groups",groupArticleGalleryImages),
 }

}

func assembleArticleDocument(extracted extraction.Result, sourceHTML string) Article {
 return assembly.Run(extracted,sourceHTML,articleAssemblyStages()...)
}
