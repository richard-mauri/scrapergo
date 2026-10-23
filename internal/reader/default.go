package reader

import (
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/contributors"
 "clean-article-reader/internal/documentmedia"
 "clean-article-reader/internal/media"
 "clean-article-reader/internal/recovery"
)

// DefaultProcessor composes the production article workflow from backend
// implementations. No callbacks from the executable package are required.
func DefaultProcessor() Processor {
 return New(Dependencies{
  RecoverPhotographs: recovery.RecoverArticlePhotographs,
  RecoverLinks: recovery.RecoverMissingArticleLinks,
  RecoverAuthorPortrait: contributors.RecoverAuthorPortrait,
  ClassifyContributors: contributors.ClassifyContributorImages,
  ClassifySourceImages: recovery.ClassifyImagesFromSource,
  ReconcileByline: contributors.ReconcileContributorByline,
  RecoverOpeningFigure: recovery.RecoverOpeningSourceFigure,
  Normalize: func(a *article.Article, sourceLead string) {
   documentmedia.NormalizeDocument(a,media.SourceIdentity,sourceLead)
  },
  RestoreGalleryGroups: recovery.GroupArticleGalleryImages,
 })
}
