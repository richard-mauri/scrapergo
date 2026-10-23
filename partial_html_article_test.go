package main

import (
 articlemodel "clean-article-reader/internal/article"
 "clean-article-reader/internal/bootstrap"
 "net/http"
 "net/http/httptest"
 "encoding/json"
 "testing"
)

const partialFixture = `<!doctype html><html><head>
<meta property="og:type" content="article">
<meta property="og:title" content="Quarterly earnings report disappoints investors">
<link rel="canonical" href="https://example.org/news/report">
<meta property="og:image" content="https://example.org/pic.jpg">
</head><body><nav><p>Home Markets Markets News</p></nav>
<article><h1>Quarterly earnings report disappoints investors</h1>
<figure><figcaption><p>Photograph by Example</p></figcaption></figure>
<p>Shares declined sharply in early trading after an earnings report missed analysts' expectations. Management explained that unexpectedly high operating expenses had pressured margins, while distributors were also reducing their inventories amid weaker demand.</p>
<p>The company said the outlook for the next quarter remains uncertain as competitors respond to shifting conditions. Details were not yet available.</p>
</article></body></html>`

func TestPartialExcerptThroughNormalExtractionPipeline(t *testing.T) {
    response := httptest.NewRecorder()
    bootstrap.WriteExtractedArticle(response, partialFixture, "https://example.org/news/report")
    if response.Code != http.StatusOK {
        t.Fatalf("partial article rejected: HTTP %d: %s", response.Code, response.Body.String())
    }
    var article articlemodel.Article
    if err := json.Unmarshal(response.Body.Bytes(), &article); err != nil { t.Fatal(err) }
    if article.Method != "Partial HTML excerpt" || article.AccessNote == "" {
        t.Fatalf("missing partial excerpt signaling: method=%q note=%q", article.Method, article.AccessNote)
    }
}
