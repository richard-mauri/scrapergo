package main
import "clean-article-reader/internal/recovery"
func recoverOpeningSourceFigure(a *Article, raw, source string) {
 recovery.RecoverOpeningSourceFigure(a,raw,source)
}
func recoverFigureInScope(a *Article, scope *Node, source string) bool {
 return recovery.RecoverFigureInScope(a,scope,source)
}
