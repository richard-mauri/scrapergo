package main

import (
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/assembly"
 "clean-article-reader/internal/extraction"
 "clean-article-reader/internal/reader"
)

// The executable package temporarily adapts historical source transforms.
// The reader owns sequencing and isolation; PR #43 relocates these callbacks.
func articleProcessor() reader.Processor {
 return reader.New(reader.Dependencies{
  RecoverPhotographs: recoverArticlePhotographs,
  RecoverLinks: recoverMissingArticleLinks,
  RecoverAuthorPortrait: recoverAuthorPortrait,
  ClassifyContributors: classifyContributorImages,
  ClassifySourceImages: classifyImagesFromSource,
  ReconcileByline: reconcileContributorByline,
  RecoverOpeningFigure: recoverOpeningSourceFigure,
  Normalize: func(a *article.Article, lead string) { normalizeArticleDocument(a, lead) },
  RestoreGalleryGroups: groupArticleGalleryImages,
 })
}

func articleAssemblyStages() []assembly.Stage { return articleProcessor().Stages() }

func assembleArticleDocument(extracted extraction.Result, sourceHTML string) article.Article {
 return articleProcessor().Process(extracted, sourceHTML)
}
