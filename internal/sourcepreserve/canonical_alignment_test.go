package sourcepreserve

import (
 "clean-article-reader/internal/article"
 "testing"
)
func TestCanonicalImageAlignmentTracksRealBlocksAndStandaloneLeadSeparately(t *testing.T) {
 doc,err:=ConvertSelectedHTML(`<article><img src="/a.jpg"><img src="/b.jpg"><img src="/c.jpg"></article>`,baseURL)
 if err!=nil{t.Fatal(err)}
 current:=article.Article{
  ImageURL:"https://example.org/hero.jpg",
  Blocks:[]article.Block{{Kind:"p",Text:"Unrelated prose"},{Kind:"img",URL:"https://example.org/a.jpg"},{Kind:"img",URL:"https://example.org/c.jpg"},{Kind:"img",URL:"https://example.org/other.jpg"}},
 }
 got:=CompareCanonicalImages(current,doc)
 if got.CanonicalImageBlocks!=3||got.SelectedImages!=3||got.ExactURLMatches!=2||got.CanonicalOnly!=1||got.SelectedOnly!=1||!got.OrderPreserved||!got.CanonicalLeadImagePresent {
  t.Fatalf("canonical comparison wrong: %+v",got)
 }
}
func TestCanonicalImageAlignmentDetectsOrderChangesAndRepeatedURLs(t *testing.T) {
 doc,err:=ConvertSelectedHTML(`<img src="/a.jpg"><img src="/b.jpg"><img src="/a.jpg">`,baseURL)
 if err!=nil{t.Fatal(err)}
 current:=article.Article{Blocks:[]article.Block{{Kind:"img",URL:"https://example.org/a.jpg"},{Kind:"img",URL:"https://example.org/a.jpg"},{Kind:"img",URL:"https://example.org/b.jpg"}}}
 got:=CompareCanonicalImages(current,doc)
 if got.ExactURLMatches!=3||got.SelectedOnly!=0||got.CanonicalOnly!=0||got.OrderPreserved {t.Fatalf("reorder / repeated URL missing: %+v",got)}
 if got.Matches[0].SourceIndex!=1||got.Matches[1].SourceIndex!=3||got.Matches[2].SourceIndex!=2{t.Fatalf("wrong mapping: %+v",got.Matches)}
}
