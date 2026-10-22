package main

import "clean-article-reader/internal/media"

// Deprecated compatibility entry points for tests and existing callers.
// The implementation and its policy now belong to internal/media.
func imageIdentity(raw string) string { return media.Identity(raw) }
func deduplicateArticleImages(a *Article) { media.Deduplicate(a) }
