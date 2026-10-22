package main

import (
    "strings"
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
<p>The company said the outlook for the next quarter remains uncertain as competitors respond to shifting conditions. Additional details were not yet available in the delivered document.</p>
</article></body></html>`

func TestPartialHTMLArticleAcceptsCorroboratedExcerpt(t *testing.T) {
    a, ok := partialHTMLArticle(partialFixture, "https://example.org/news/report")
    if !ok {
        t.Fatal("expected a strongly evidenced short article excerpt")
    }
    if a.WordCount < 30 || a.WordCount >= 60 || len(a.Blocks) != 2 {
        t.Fatalf("incorrect partial extraction: words=%d blocks=%d", a.WordCount, len(a.Blocks))
    }
    if strings.Contains(a.Blocks[0].Text, "Photograph") {
        t.Fatal("figure credit leaked into prose")
    }
    if a.AccessNote == "" { t.Fatal("partial result must be labeled") }
}

func TestPartialHTMLArticleRejectsIndexAndUnmatchedCanonical(t *testing.T) {
    if _, ok := partialHTMLArticle(strings.ReplaceAll(partialFixture, "content=\"article\"", "content=\"website\""), "https://example.org/news/report"); ok {
        t.Fatal("index should not qualify")
    }
    if _, ok := partialHTMLArticle(partialFixture, "https://example.org/news/different"); ok {
        t.Fatal("canonical mismatch should disqualify excerpt")
    }
}
