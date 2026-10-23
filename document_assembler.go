package main

import (
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/assembly"
 "clean-article-reader/internal/extraction"
 "clean-article-reader/internal/reader"
)

// Compatibility entry points for existing root-package tests and extraction.
// The production stage graph is now exclusively defined in internal/reader.
func articleAssemblyStages() []assembly.Stage { return reader.DefaultProcessor().Stages() }

func assembleArticleDocument(result extraction.Result, sourceHTML string) article.Article {
 return reader.DefaultProcessor().Process(result,sourceHTML)
}
