package main

import (
 "clean-article-reader/internal/sourcehtml"
 "net/url"

 extractionmodel "clean-article-reader/internal/extraction"
)

// extractWithEvidence preserves the initial extractor result and independently
// records the publisher's source metadata. No later normalization decision is
// made while gathering this evidence.
func extractWithEvidence(raw, source string) (extractionmodel.Result, error) {
 article, err := extractionmodel.ExtractWithReadability(raw, source, extractionmodel.LegacyExtract, extractionmodel.InlineTextParts)
 if err != nil { return extractionmodel.Result{}, err }
 result := extractionmodel.Result{Article: article, Evidence: extractionmodel.SourceEvidence{SourceURL: source}}
 if base, err := url.Parse(source); err == nil {
  result.Evidence.MetadataImageURL = sourcehtml.ResolveURL(meta(parseHTML(raw), "og:image", "twitter:image"), base)
 }
 return result, nil
}
