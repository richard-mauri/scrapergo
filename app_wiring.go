package main

import (
 "net/http"

 "clean-article-reader/internal/httpapi"
)

// newArticleReadHandler is the single production dependency wiring point for
// the primary read endpoint. It deliberately reuses existing security,
// retrieval and extraction policies instead of recreating them in HTTP code.
func newArticleReadHandler() http.Handler {
 return httpapi.ReadHandler(httpapi.ReadDependencies{
  Resolve: resolveGoogleNewsURL,
  Download: download,
  PublisherAccessMessage: publisherAccessMessage,
  Hostname: mustHostname,
  WriteExtracted: writeExtractedArticle,
 })
}

// newArticleMetadataHandler shares the exact read handler instance provided by
// the application's router, preventing mismatched access policies.
func newArticleMetadataHandler(read http.Handler) http.Handler {
 return httpapi.ImageMetadataHandler(read)
}
