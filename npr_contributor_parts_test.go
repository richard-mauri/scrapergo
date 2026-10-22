package main

import "testing"

func TestNPRContributorPartsHeaderSeparation(t *testing.T) {
 name:=func(s string) Block{return Block{Kind:"p",Text:s,Parts:[]TextPart{{Text:s,URL:"https://www.npr.org/people/example"}}}}
 a:=Article{Byline:"Ruth Sherlock",Blocks:[]Block{
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
 separateContributorHeader(&a)
 if a.Byline!="Ruth Sherlock, Samantha Balaban, Nuha Musleh, Daniel Estrin" {t.Fatalf("wrong byline %q",a.Byline)}
 if len(a.Blocks)!=6 {t.Fatalf("unexpected remaining blocks %+v",a.Blocks)}
 if a.Blocks[0].Text!="Middle East"||a.Blocks[1].Text!="Brad Lander West Bank"||a.Blocks[2].Text!="Download"||a.Blocks[3].URL!=nprWestBankFirst||a.Blocks[5].URL!=nprWestBankInline{t.Fatalf("editorial body damaged %+v",a.Blocks)}
}
func TestNoBylineHeaderDoesNotConsumeEditorialLinks(t *testing.T){
 a:=Article{Blocks:[]Block{{Kind:"p",Text:"Opening words"},{Kind:"p",Text:"Linked article",Parts:[]TextPart{{Text:"Linked article",URL:"https://example.org"}}},{Kind:"img",Role:"editorial",URL:"https://example.org/scene.jpg"}}}
 separateContributorHeader(&a)
 if len(a.Blocks)!=3 {t.Fatalf("unexpected removal %+v",a.Blocks)}
}
