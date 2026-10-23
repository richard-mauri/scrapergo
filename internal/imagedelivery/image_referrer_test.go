package imagedelivery

import (

    "net/url"
    "testing"
)

func TestArticleReferrerRetainsPathWithoutSensitiveQuery(t *testing.T) {
    u, err := url.Parse("https://www.example.org/news/story?utm_source=rss&token=secret#gallery")
    if err != nil { t.Fatal(err) }
    got := ArticleReferrer(u)
    want := "https://www.example.org/news/story"
    if got != want { t.Fatalf("referrer = %q, want %q", got, want) }
    if u.RawQuery == "" || u.Fragment == "" { t.Fatal("article URL was modified") }
}

func TestArticleReferrerPreservesEncodedPath(t *testing.T) {
    u, err := url.Parse("https://www.example.org/articles/a%20b?source=feed")
    if err != nil { t.Fatal(err) }
    if got := ArticleReferrer(u); got != "https://www.example.org/articles/a%20b" {
        t.Fatalf("unexpected referrer %q", got)
    }
}
