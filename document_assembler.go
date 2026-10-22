package main

import extractionmodel "clean-article-reader/internal/extraction"

// assembleArticleDocument is the sole production orchestration point for
// article reconciliation. Transform order is intentionally unchanged: media
// recovery, link repair, contributor identification, and normalization are
// order-sensitive in the v0.2.85 article contract.
//
// This initial boundary does not yet isolate individual transformations into
// packages. Their ownership can be migrated independently once representative
// source fixtures have been frozen.
func assembleArticleDocument(extracted extractionmodel.Result, sourceHTML string) Article {
 article := extracted.Article
 sourceURL := extracted.Evidence.SourceURL
 recoverArticlePhotographs(&article, sourceHTML, sourceURL)
 deduplicateArticleImages(&article)
 recoverMissingArticleLinks(&article, sourceHTML, sourceURL)
 recoverAuthorPortrait(&article, sourceHTML, sourceURL)
 removeSerializedPlayerBlocks(&article)
 classifyContributorImages(&article)
 classifyImagesFromSource(&article, sourceHTML, sourceURL)
 reconcileContributorByline(&article)
 separateContributorHeader(&article)
 recoverOpeningSourceFigure(&article, sourceHTML, sourceURL)
 normalizeArticleDocument(&article, extracted.Evidence.MetadataImageURL)
 return article
}
