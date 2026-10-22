package main

import (
 "net/http"
 "clean-article-reader/internal/httpapi"
)

// Compatibility entry point, with existing extraction and access policy
// injected rather than duplicated in the HTTP presentation package.
func imageMetadata(w http.ResponseWriter,r *http.Request) {
 newArticleMetadataHandler(newArticleReadHandler()).ServeHTTP(w,r)
}

func articleImageCandidates(a Article) []string {
 return httpapi.ImageCandidates(a)
}
