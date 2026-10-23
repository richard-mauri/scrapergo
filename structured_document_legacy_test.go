package main

import (
 "clean-article-reader/internal/contributors"
 "clean-article-reader/internal/documentmedia"
 "clean-article-reader/internal/media"
)

func imageSourceIdentity(raw string) string { return media.SourceIdentity(raw) }
func classifyContributorImages(a *Article) { contributors.ClassifyContributorImages(a) }
func normalizeArticleDocument(a *Article, sourceLead ...string) {
 documentmedia.NormalizeDocument(a,media.SourceIdentity,sourceLead...)
}
func recoverAuthorPortrait(a *Article, html, source string) {
 contributors.RecoverAuthorPortrait(a,html,source)
}
func openingContributorGroup(n *Node) bool { return contributors.OpeningContributorGroup(n) }
func reconcileContributorByline(a *Article) { contributors.ReconcileContributorByline(a) }
