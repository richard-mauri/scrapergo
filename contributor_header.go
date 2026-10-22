package main

import "clean-article-reader/internal/contributors"

// Compatibility wrappers for legacy callers and regression tests.
func linkedContributorName(b Block) bool { return contributors.LinkedName(b) }
func separateContributorHeader(a *Article) { contributors.SeparateHeader(a) }
