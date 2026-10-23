package main

import (
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/assembly"
 "clean-article-reader/internal/extraction"
 "clean-article-reader/internal/reader"
 "clean-article-reader/internal/recovery"
)

// The executable package temporarily adapts historical source transforms.
// The reader owns sequencing and isolation; PR #43 relocates these callbacks.
func articleProcessor() reader.Processor {
 return reader.New(reader.Dependencies{
  RecoverPhotographs: recoverArticlePhotographs,
  RecoverLinks: recovery.RecoverMissingArticleLinks,
  RecoverAuthorPortrait: recoverAuthorPortrait,
  ClassifyContributors: classifyContributorImages,
  ClassifySourceImages: func(a *article.Article, raw, u string) { recovery.ClassifyImagesFromSource(a,raw,u,articleImageURL,imageSourceIdentity) },
  ReconcileByline: reconcileContributorByline,
  RecoverOpeningFigure: func(a *article.Article, raw, u string) { recovery.RecoverOpeningSourceFigure(a,raw,u,articleImageURL,imageSourceIdentity) },
  Normalize: func(a *article.Article, lead string) { normalizeArticleDocument(a, lead) },
  RestoreGalleryGroups: func(a *article.Article, raw, u string) { recovery.GroupArticleGalleryImages(a,raw,u,articleImageURL) },
 })
}

func articleAssemblyStages() []assembly.Stage { return articleProcessor().Stages() }

func assembleArticleDocument(extracted extraction.Result, sourceHTML string) article.Article {
 return articleProcessor().Process(extracted, sourceHTML)
}
