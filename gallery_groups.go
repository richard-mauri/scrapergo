package main
import (
 "clean-article-reader/internal/recovery"
)
func sourceGalleryCandidate(n *Node) bool {return recovery.SourceGalleryCandidate(n)}
func groupArticleGalleryImages(a *Article, raw, source string) {
 recovery.GroupArticleGalleryImages(a,raw,source,articleImageURL)
}
