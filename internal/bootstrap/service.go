package bootstrap

import (
 "clean-article-reader/internal/accesspolicy"
 "clean-article-reader/internal/app"
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/diagnostics"
 "clean-article-reader/internal/extraction"
 "clean-article-reader/internal/fetcher"
 "clean-article-reader/internal/googlenews"
 "clean-article-reader/internal/httpapi"
 "clean-article-reader/internal/imagedelivery"
 "clean-article-reader/internal/pageclassify"
 "clean-article-reader/internal/pdfarticle"
 "clean-article-reader/internal/reader"
 "clean-article-reader/internal/sourcehtml"
 "clean-article-reader/internal/structured"
 "context"
 "encoding/json"
 "log"
 "net/http"
 "net/url"
 "strings"
)
// writeExtractedArticle is shared by normal retrieval and diagnostic HTML submission.
func WriteExtractedArticle(w http.ResponseWriter, raw, finalURL string) {
 writeExtractedArticle(context.Background(),w,raw,finalURL,false)
}

// WriteHTMLWithPDF is the network read path. Diagnostics deliberately retain
// the pure offline HTML extractor without PDF subrequests.
func WriteHTMLWithPDF(ctx context.Context, w http.ResponseWriter, raw, finalURL string) {
 writeExtractedArticle(ctx,w,raw,finalURL,true)
}

func writeExtractedArticle(ctx context.Context, w http.ResponseWriter, raw, finalURL string, allowPDF bool) {
	LogSlateRetrievalDiagnostic(raw, finalURL)
	// A navigation index is not a single article. Classify conservatively before
	// Readability can flatten its menu and story cards into a false article.
	if (pageclassify.IsIndex(raw) || pageclassify.IsLanding(raw)) && !extraction.HasPartialArticleEvidence(raw, finalURL) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"kind":  "page_index",
			"error": "This appears to be a homepage or index, not a single article. Open the original website to browse its links.",
			"url":   finalURL,
		})
		return
	}
	// Extract before deciding whether restricted pages offer readable public text.
	// Restriction metadata alone does not establish that no usable preview exists.
	kind := accesspolicy.Restriction(raw)
	cleanedArticleHTML := structured.StripEmbeddedMedia(raw)
	extracted, extractionErr := ExtractWithEvidence(cleanedArticleHTML, finalURL)
	article := extracted.Article
	if extractionErr == nil {
		article = reader.DefaultProcessor().Process(extracted, cleanedArticleHTML)
		LogSlateExtractionDiagnostic(article.Blocks, finalURL)
	}
	if kind != "" {
		log.Printf("restriction publisher=%s kind=%s evidence=%s html_bytes=%d public_words=%d", accesspolicy.Hostname(finalURL), kind, accesspolicy.Evidence(raw), len(raw), article.WordCount)
		if extractionErr == nil && article.WordCount >= 80 {
			article.AccessType = kind
			article.AccessNote = accesspolicy.Notice(kind)
			_ = json.NewEncoder(w).Encode(article)
			return
		}
		w.WriteHeader(422)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": accesspolicy.Notice(kind), "kind": kind, "publisher": accesspolicy.Hostname(finalURL), "evidence": accesspolicy.Evidence(raw)})
		return
	}
	if (extractionErr == nil && (pageclassify.IsThinPortal(raw, article.WordCount) || pageclassify.IsSparseRootLanding(raw, finalURL, article.WordCount)) || extractionErr != nil && pageclassify.IsSearchPortal(raw)) && !extraction.HasPartialArticleEvidence(raw, finalURL) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"kind":  "page_index",
			"error": "This appears to be a homepage or index, not a single article. Open the original website to browse its links.",
			"url":   finalURL,
		})
		return
	}
	if extractionErr != nil && allowPDF {
        if candidate := extraction.DiscoverPDFLink(raw, finalURL); candidate != "" {
            bytes, pdfURL, err := fetcher.DownloadPDF(ctx, candidate)
            if err != nil {
                w.WriteHeader(http.StatusUnprocessableEntity)
                _ = json.NewEncoder(w).Encode(map[string]string{"kind":"pdf_download_failed","error":err.Error(),"url":finalURL})
                return
            }
            pdfArticle, err := pdfarticle.Extract(ctx,bytes,pdfURL)
            if err != nil {
                w.WriteHeader(http.StatusUnprocessableEntity)
                _ = json.NewEncoder(w).Encode(map[string]string{"kind":"pdf_extraction_failed","error":err.Error(),"url":finalURL})
                return
            }
            // Preserve the article address supplied by the user.
            pdfArticle.URL = finalURL
            pdfArticle.PDFURL = pdfURL
            _ = json.NewEncoder(w).Encode(pdfArticle)
            return
        }
    }
	if extractionErr != nil {
		w.WriteHeader(422)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "ScraperGo could not identify a readable article on this page. Open the original website to continue.", "kind": "extraction_error", "url": finalURL})
		return
	}
	_ = json.NewEncoder(w).Encode(article)
}

// extractWithEvidence preserves the initial extractor result and independently
// records the publisher's source metadata. No later normalization decision is
// made while gathering this evidence.
func ExtractWithEvidence(raw, source string) (extraction.Result, error) {
 article, err := extraction.ExtractWithReadability(raw, source, extraction.LegacyExtract, extraction.InlineTextParts)
 if err != nil { return extraction.Result{}, err }
 result := extraction.Result{Article: article, Evidence: extraction.SourceEvidence{SourceURL: source}}
 if base, err := url.Parse(source); err == nil {
  result.Evidence.MetadataImageURL = sourcehtml.ResolveURL(sourcehtml.Meta(sourcehtml.Parse(raw), "og:image", "twitter:image"), base)
 }
 return result, nil
}
// Diagnostic only: no raw article contents are logged and no extraction changes.
func SlateDiagnosticTarget(source string) bool {
	u, err := url.Parse(source)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Hostname(), "arstechnica.com") &&
		strings.Contains(u.Path, "/visiting-slates-first-showroom-as-its-25000-electric-truck-nears-production/")
}

func LogSlateRetrievalDiagnostic(raw, source string) {
	if !SlateDiagnosticTarget(source) {
		return
	}
	// Raw markers may come from JSON/script content; visible markers cannot.
	visible := strings.ToLower(sourcehtml.Text(sourcehtml.Parse(raw)))
	lower := strings.ToLower(raw)
	log.Printf("slate_diagnostic stage=download html_bytes=%d raw_slate_garage=%t visible_slate_garage=%t raw_repairpal=%t visible_repairpal=%t raw_closing=%t visible_closing=%t raw_pegboard=%t raw_interior_models=%t",
		len(raw),
		strings.Contains(lower, "beyond the default blank slate body style"),
		strings.Contains(visible, "beyond the default blank slate body style"),
		strings.Contains(lower, "for service, slate will be using"),
		strings.Contains(visible, "for service, slate will be using"),
		strings.Contains(lower, "now it’s time to deliver") || strings.Contains(lower, "now it's time to deliver"),
		strings.Contains(visible, "now it’s time to deliver") || strings.Contains(visible, "now it's time to deliver"),
		strings.Contains(lower, "slate-visit-chicago-pegboard"),
		strings.Contains(lower, "slate-visit-chicago-interior-models"),
	)
}

func LogSlateExtractionDiagnostic(blocks []article.Block, source string) {
	if !SlateDiagnosticTarget(source) {
		return
	}
	var texts []string
	images := 0
	for _, b := range blocks {
		texts = append(texts, b.Text)
		if b.Kind == "img" {
			images++
			texts = append(texts, b.URL)
		}
	}
	value := strings.ToLower(strings.Join(texts, " "))
	log.Printf("slate_diagnostic stage=extraction blocks=%d images=%d slate_garage=%t repairpal=%t closing=%t pegboard=%t interior_models=%t",
		len(blocks), images,
		strings.Contains(value, "beyond the default blank slate body style"),
		strings.Contains(value, "for service, slate will be using"),
		strings.Contains(value, "now it’s time to deliver") || strings.Contains(value, "now it's time to deliver"),
		strings.Contains(value, "slate-visit-chicago-pegboard"),
		strings.Contains(value, "slate-visit-chicago-interior-models"),
	)
}
// newArticleReadHandler is the single production dependency wiring point for
// the primary read endpoint. It deliberately reuses existing security,
// retrieval and extraction policies instead of recreating them in HTTP code.
func NewArticleReadHandler() http.Handler {
 return httpapi.ReadHandler(httpapi.ReadDependencies{
  Resolve: googlenews.ResolveURL,
  Download: fetcher.Download,
  PublisherAccessMessage: fetcher.PublisherAccessMessage,
  Hostname: accesspolicy.Hostname,
  WriteExtracted: WriteExtractedArticle,
  WriteExtractedWithContext: WriteHTMLWithPDF,
 })
}

// newArticleMetadataHandler shares the exact read handler instance provided by
// the application's router, preventing mismatched access policies.
func NewArticleMetadataHandler(read http.Handler) http.Handler {
 return httpapi.ImageMetadataHandler(read)
}

// Start composes the production HTTP graph once.
func Start(port string, version, build string, tannerJPG []byte) error {
 readHandler := NewArticleReadHandler()
 return app.Run(app.Config{Port:port,Version:version,Build:build}, app.Handlers{
  Read:readHandler,
  Classify:(&httpapi.Classifier{Read:readHandler,ValidateURL:fetcher.ValidatedURL}).ServeHTTP,
  ImageMetadata:NewArticleMetadataHandler(readHandler),
  ArticleImage:(imagedelivery.Handler{Metadata:NewArticleMetadataHandler(readHandler)}).ServeHTTP,
  PDF:httpapi.PDFHandler(fetcher.DownloadPDF),
  Thumbnail:httpapi.NewThumbnailHandler(fetcher.ValidatedURL,imagedelivery.Handler{Metadata:NewArticleMetadataHandler(readHandler)}).ServeHTTP,
  DiagnosticExtract:(diagnostics.HTMLHandler{WriteExtracted:WriteExtractedArticle}).ServeHTTP,
  DiagnosticRawHTML:httpapi.RawHTMLDiagnosticHandler(fetcher.Download),
  DiagnosticImageTrace:(diagnostics.ImageTraceHandler{Extract:ExtractWithEvidence}).ServeHTTP,
  DiagnosticClassificationTrace:classificationTraceHandler(readHandler),
  TannerPhoto:tannerJPG,
 })
}
