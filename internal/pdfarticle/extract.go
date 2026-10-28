// Package pdfarticle converts text-based PDFs to ScraperGo's canonical article model.
// No publisher network access is permitted from this package.
package pdfarticle

import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "net/url"
    "path"
    "strings"

    "clean-article-reader/internal/article"
    pdf "github.com/giraffesyo/pdf"
)

const MaxPDFBytes = 12 << 20
const MaxPDFPages = 100
const MaxTextBytes = 2 << 20

var ErrNoReadableText = errors.New("PDF contains no extractable text (it may be scanned)")
var ErrPDFTooLarge = errors.New("PDF exceeds size limit")
var ErrPDFTooManyPages = errors.New("PDF exceeds page limit")

// Extract never fetches URLs. It only interprets bytes already obtained through
// the validated fetcher, and bounds both retained output and page count.
func Extract(ctx context.Context, data []byte, sourceURL string) (article.Article, error) {
    if len(data) == 0 || len(data) > MaxPDFBytes { return article.Article{}, ErrPDFTooLarge }
    if !bytes.HasPrefix(data, []byte("%PDF-")) { return article.Article{}, errors.New("response is not a PDF document") }
    u, err := url.Parse(sourceURL)
    if err != nil || (u.Scheme != "http" && u.Scheme != "https") { return article.Article{}, errors.New("invalid PDF source URL") }
    document, err := pdf.Extract(ctx, bytes.NewReader(data), int64(len(data)))
    if err != nil { return article.Article{}, fmt.Errorf("PDF extraction failed: %w", err) }
    if document.PageCount > MaxPDFPages { return article.Article{}, ErrPDFTooManyPages }
    a := article.Article{URL: sourceURL, Site: u.Hostname(), Title: strings.TrimSuffix(path.Base(u.Path), path.Ext(u.Path)), Method: "PDF text extraction"}
    if a.Title == "" || a.Title == "." || a.Title == "/" { a.Title = "PDF article" }
    total := 0
    for _, page := range document.Pages {
        if err := ctx.Err(); err != nil { return article.Article{}, err }
        for _, paragraph := range strings.Split(strings.TrimSpace(page.Text()), "\n\n") {
            paragraph = strings.TrimSpace(paragraph)
            if paragraph == "" { continue }
            total += len(paragraph)
            if total > MaxTextBytes { return article.Article{}, errors.New("PDF extracted text exceeds limit") }
            a.Blocks = append(a.Blocks, article.Block{Kind:"p", Text:paragraph})
            a.WordCount += len(strings.Fields(paragraph))
        }
    }
    if a.WordCount < 20 { return article.Article{}, ErrNoReadableText }
    return a, nil
}
