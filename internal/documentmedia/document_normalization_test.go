package documentmedia

import (
 "testing"
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/media"
)

func TestNormalizeArticleDocumentTitleAndImage(t *testing.T) {
 a:=article.Article{Title:"News headline",ImageURL:"https://example.org/hero.jpg",Blocks:[]article.Block{{Kind:"h2",Text:"News headline"},{Kind:"p",Text:"Paragraph"},{Kind:"img",URL:"https://example.org/hero.jpg"},{Kind:"img",URL:"https://example.org/portrait.jpg"}}}
 NormalizeDocument(&a,media.SourceIdentity)
 if len(a.Blocks)!=3 || a.Blocks[0].Text!="Paragraph"{t.Fatalf("opening title handling: %+v",a.Blocks)}
 if a.ImageURL!="https://example.org/hero.jpg" || a.Blocks[1].Role!="lead" {t.Fatalf("lead identity or metadata lost: %+v",a)}
 if a.Blocks[2].URL!="https://example.org/portrait.jpg"{t.Fatal("portrait lost")}
}

func TestNormalizeArticleDocumentDistinctHero(t *testing.T) {
 a:=article.Article{Title:"News headline",ImageURL:"https://example.org/lead.jpg",Blocks:[]article.Block{{Kind:"img",Role:"author",URL:"https://example.org/author.jpg"},{Kind:"p",Text:"Editorial paragraph"}}}
 NormalizeDocument(&a,media.SourceIdentity)
 if len(a.Blocks)!=3 || a.Blocks[0].Role!="lead" || a.Blocks[0].URL!="https://example.org/lead.jpg" || a.Blocks[1].URL!="https://example.org/author.jpg" {t.Fatalf("distinct lead or author image missing: %+v",a.Blocks)}
}


func TestSourceMetadataMarksMatchingInlineLead(t *testing.T) {
 a:=article.Article{Title:"Story",Blocks:[]article.Block{{Kind:"img",Role:"author",URL:"https://example.org/author.jpg"},{Kind:"img",URL:"https://example.org/cover.jpg"}}}
 NormalizeDocument(&a,media.SourceIdentity,"https://example.org/cover.jpg")
 if len(a.Blocks)!=2 || a.Blocks[0].Role=="lead" || a.Blocks[1].Role!="lead" {t.Fatalf("lead duplicated or misidentified: %+v",a.Blocks)}
}
