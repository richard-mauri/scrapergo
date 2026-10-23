package main
import "clean-article-reader/internal/recovery"
func recoverMissingArticleLinks(a *Article, raw, source string) {
 recovery.RecoverMissingArticleLinks(a,raw,source)
}
