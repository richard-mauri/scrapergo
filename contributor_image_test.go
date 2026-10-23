package main

import (
 articlemodel "clean-article-reader/internal/article"
 "testing"
 "clean-article-reader/internal/contributors"
 "clean-article-reader/internal/documentmedia"
 "clean-article-reader/internal/media"
)

const estrinPortrait = "https://media.npr.org/assets/img/2020/01/02/daniel-estrin_sq-4199304dee8cdf91935a65a2d0b069e07311386e.jpg?s=100&c=85&f=jpeg"

func TestContributorPortraitNotEditorialLead(t *testing.T) {
 a:=articlemodel.Article{ImageURL:nprWestBankMetadata,Blocks:[]articlemodel.Block{
  {Kind:"img",Role:"author",URL:"https://example.org/samantha.png",Text:"Photo of Samantha Balaban in the field"},
  {Kind:"p",Text:"Daniel Estrin",Links:[]articlemodel.InlineLink{{Text:"Daniel Estrin",URL:"https://www.npr.org/people/123"}}},
  {Kind:"img",Text:"Daniel Estrin",URL:estrinPortrait},
  {Kind:"p",Text:"Daniel Estrin"},
  {Kind:"img",Role:"lead",URL:nprWestBankFirst,Text:"Israeli security forces"},
  {Kind:"p",Text:"Story reporting"},
  {Kind:"img",URL:nprWestBankInline,Text:"Activists and rabbis from the United States and Israel help Palestinians harvest olives"},
 }}
 contributors.ClassifyContributorImages(&a)
 if a.Blocks[2].Role!="author"{t.Fatalf("contributor portrait not classified: %+v",a.Blocks[2])}
 documentmedia.NormalizeDocument(&a,media.SourceIdentity,nprWestBankMetadata)
 if a.ImageURL!=nprWestBankFirst {t.Fatalf("wrong article lead %s",a.ImageURL)}
 if a.Blocks[2].Role!="author" || a.Blocks[4].Role!="lead" || a.Blocks[6].URL!=nprWestBankInline {t.Fatalf("image roles or order wrong: %+v",a.Blocks)}
}
func TestEditorialCaptionNotPromotedToContributor(t *testing.T) {
 a:=articlemodel.Article{Blocks:[]articlemodel.Block{{Kind:"p",Text:"Photo by Daniel Estrin"},{Kind:"img",Text:"Military patrols in the area",URL:nprWestBankFirst}}}
 contributors.ClassifyContributorImages(&a)
 if a.Blocks[1].Role=="author"{t.Fatal("editorial figure incorrectly classified")}
}
