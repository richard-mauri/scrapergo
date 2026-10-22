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
 if len(a.Blocks)!=4 || a.Blocks[1].URL!="https://cdn.example.org/hero.jpg" || a.Blocks[2].Text!="Photo caption" || a.Blocks[3].Text!="Closing text" {t.Fatalf("unexpected blocks: %+v",a.Blocks)}
}

func TestDeduplicateNilArticle(t *testing.T) { Deduplicate(nil) }

func TestDeduplicateKeepsBMWLeadWithResizedMetadata(t *testing.T) {
 const prefix="https://cdn.arstechnica.net/wp-content/uploads/2026/10/"
 lead:=prefix+"BMW-iX5-60-xDrive-Tanzanite-Blue_43.jpg"
 a:=article.Article{ImageURL:prefix+"BMW-iX5-60-xDrive-Tanzanite-Blue_43-1152x648.jpg",Blocks:[]article.Block{
  {Kind:"p",Text:"Opening"},
  {Kind:"img",URL:lead,Text:"BMW X5 daylight running lights"},
  {Kind:"p",Text:"Body"},
  {Kind:"img",URL:prefix+"BMW-iX5-60-xDrive-Pine-Grey_23-1024x683.jpg",Text:"BMW drives on the road"},
  {Kind:"img",URL:prefix+"BMW-iX5-60-xDrive-Tanzanite-Blue_43-640x360.jpg",Text:"Duplicate lead rendition"},
 }}
 Deduplicate(&a)
 if len(a.Blocks)!=4 || a.Blocks[1].URL!=lead || a.Blocks[3].Text!="BMW drives on the road" {t.Fatalf("lost opening editorial figure or gallery: %+v",a.Blocks)}
}

func TestDeduplicateDistinctMetadataNeverDeletesInline(t *testing.T) {
 a:=article.Article{ImageURL:"https://example.org/metadata.jpg",Blocks:[]article.Block{
 {Kind:"img",URL:"https://example.org/lead.jpg"},
 {Kind:"img",URL:"https://example.org/gallery.jpg"},
 }}
 Deduplicate(&a)
 if len(a.Blocks)!=2 {t.Fatalf("unexpected removal: %+v",a.Blocks)}
}
