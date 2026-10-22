package documentmedia

import (
 "strings"
 "testing"
 "clean-article-reader/internal/article"
)

func identity(s string)string{return strings.Split(s,"?")[0]}

func TestNormalizeDocumentPrioritizesEditorialImage(t *testing.T){
 a:=article.Article{Title:"A reported article",ImageURL:"https://cdn.example.test/preview.jpg",Blocks:[]article.Block{
  {Kind:"h1",Text:"A reported article"},
  {Kind:"img",URL:"https://cdn.example.test/portrait.jpg",Role:"author"},
  {Kind:"img",URL:"https://cdn.example.test/editorial.jpg?size=small",Role:"editorial",Text:"First figure"},
  {Kind:"img",URL:"https://cdn.example.test/editorial.jpg?size=large",Text:"Another caption"},
 }}
 NormalizeDocument(&a,identity)
 if a.ImageURL!="https://cdn.example.test/editorial.jpg?size=small"{t.Fatalf("wrong lead: %s",a.ImageURL)}
 if len(a.Blocks)!=2 || a.Blocks[0].Role!="author" || a.Blocks[1].Role!="lead" || a.Blocks[1].Text!="First figure" {t.Fatalf("unexpected blocks: %+v",a.Blocks)}
}

func TestNormalizeDocumentMetadataRenditionAndFallback(t *testing.T){
 a:=article.Article{Blocks:[]article.Block{{Kind:"img",URL:"https://cdn.example.test/photo.jpg?small",Role:"editorial"}}}
 NormalizeDocument(&a,identity,"https://cdn.example.test/photo.jpg?large")
 if a.ImageURL!="https://cdn.example.test/photo.jpg?large" || a.Blocks[0].URL!=a.ImageURL {t.Fatalf("metadata rendition not reused: %+v",a)}
 b:=article.Article{Blocks:[]article.Block{{Kind:"p",Text:"Story text"}}}
 NormalizeDocument(&b,identity,"https://cdn.example.test/preview.jpg")
 if len(b.Blocks)!=2 || b.Blocks[0].Role!="lead" || b.ImageURL!="https://cdn.example.test/preview.jpg" {t.Fatalf("fallback missing: %+v",b)}
}
