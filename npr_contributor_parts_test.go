package main

import (
 articlemodel "clean-article-reader/internal/article"
 "testing"
"clean-article-reader/internal/contributors"
)

func TestNPRContributorPartsHeaderSeparation(t *testing.T) {
 name:=func(s string) articlemodel.Block{return articlemodel.Block{Kind:"p",Text:s,Parts:[]articlemodel.TextPart{{Text:s,URL:"https://www.npr.org/people/example"}}}}
 a:=articlemodel.Article{Byline:"Ruth Sherlock",Blocks:[]articlemodel.Block{
  {Kind:"p",Text:"Middle East"},{Kind:"p",Text:"By"},
  name("Ruth Sherlock"),{Kind:"p",Text:","},
  {Kind:"img",Role:"author",URL:"https://example.org/samantha.png"},
  name("Samantha Balaban"),{Kind:"p",Text:","},
  name("Nuha Musleh"),{Kind:"p",Text:","},
  {Kind:"img",Role:"author",URL:estrinPortrait},
  name("Daniel Estrin"),
  {Kind:"h3",Text:"Brad Lander West Bank"},
  {Kind:"li",Text:"Download"},
  {Kind:"img",Role:"editorial",URL:nprWestBankFirst,Text:"Israeli security forces"},
  {Kind:"p",Text:"Article begins here."},
  {Kind:"img",Role:"editorial",URL:nprWestBankInline,Text:"Olive harvesting"},
 }}
 contributors.SeparateHeader(&a)
 if a.Byline!="Ruth Sherlock, Samantha Balaban, Nuha Musleh, Daniel Estrin" {t.Fatalf("wrong byline %q",a.Byline)}
 if len(a.Blocks)!=6 {t.Fatalf("unexpected remaining blocks %+v",a.Blocks)}
 if a.Blocks[0].Text!="Middle East"||a.Blocks[1].Text!="Brad Lander West Bank"||a.Blocks[2].Text!="Download"||a.Blocks[3].URL!=nprWestBankFirst||a.Blocks[5].URL!=nprWestBankInline{t.Fatalf("editorial body damaged %+v",a.Blocks)}
}
func TestNoBylineHeaderDoesNotConsumeEditorialLinks(t *testing.T){
 a:=articlemodel.Article{Blocks:[]articlemodel.Block{{Kind:"p",Text:"Opening words"},{Kind:"p",Text:"Linked article",Parts:[]articlemodel.TextPart{{Text:"Linked article",URL:"https://example.org"}}},{Kind:"img",Role:"editorial",URL:"https://example.org/scene.jpg"}}}
 contributors.SeparateHeader(&a)
 if len(a.Blocks)!=3 {t.Fatalf("unexpected removal %+v",a.Blocks)}
}
