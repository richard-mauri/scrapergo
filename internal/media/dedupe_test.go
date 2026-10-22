package media

import (
 "testing"
 "clean-article-reader/internal/article"
)

func TestIdentityRetainsDistinctImages(t *testing.T) {
 a:=Identity("https://CDN.example.org/photos/lead-1200x800.jpg?width=1200")
 b:=Identity("https://cdn.example.org/photos/lead-600x400.jpg?width=600")
 if a=="" || a!=b {t.Fatalf("renditions differ: %q vs %q",a,b)}
 if a==Identity("https://cdn.example.org/photos/different.jpg") {t.Fatal("distinct photographs merged")}
}

func TestDeduplicateKeepsHeroAndNonImages(t *testing.T) {
 a:=article.Article{ImageURL:"https://cdn.example.org/hero.jpg",Blocks:[]article.Block{
  {Kind:"p",Text:"Opening text"},
  {Kind:"img",URL:"https://cdn.example.org/hero.jpg"},
  {Kind:"img",URL:"https://cdn.example.org/inline-800x600.jpg",Text:"Photo caption"},
  {Kind:"img",URL:"https://cdn.example.org/inline-400x300.jpg",Text:"Duplicate"},
  {Kind:"p",Text:"Closing text"},
 }}
 Deduplicate(&a)
 if len(a.Blocks)!=3 || a.Blocks[1].Text!="Photo caption" || a.Blocks[2].Text!="Closing text" {t.Fatalf("unexpected blocks: %+v",a.Blocks)}
}

func TestDeduplicateNilArticle(t *testing.T) { Deduplicate(nil) }
