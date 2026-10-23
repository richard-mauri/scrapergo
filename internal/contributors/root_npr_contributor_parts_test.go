package contributors_test

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
  {Kind:"img",Role:"author",URL:fxnprcontributorpartsestrinPortrait},
  name("Daniel Estrin"),
  {Kind:"h3",Text:"Brad Lander West Bank"},
  {Kind:"li",Text:"Download"},
  {Kind:"img",Role:"editorial",URL:fxnprcontributorpartsnprWestBankFirst,Text:"Israeli security forces"},
  {Kind:"p",Text:"Article begins here."},
  {Kind:"img",Role:"editorial",URL:fxnprcontributorpartsnprWestBankInline,Text:"Olive harvesting"},
 }}
 contributors.SeparateHeader(&a)
 if a.Byline!="Ruth Sherlock, Samantha Balaban, Nuha Musleh, Daniel Estrin" {t.Fatalf("wrong byline %q",a.Byline)}
 if len(a.Blocks)!=6 {t.Fatalf("unexpected remaining blocks %+v",a.Blocks)}
 if a.Blocks[0].Text!="Middle East"||a.Blocks[1].Text!="Brad Lander West Bank"||a.Blocks[2].Text!="Download"||a.Blocks[3].URL!=fxnprcontributorpartsnprWestBankFirst||a.Blocks[5].URL!=fxnprcontributorpartsnprWestBankInline{t.Fatalf("editorial body damaged %+v",a.Blocks)}
}
func TestNoBylineHeaderDoesNotConsumeEditorialLinks(t *testing.T){
 a:=articlemodel.Article{Blocks:[]articlemodel.Block{{Kind:"p",Text:"Opening words"},{Kind:"p",Text:"Linked article",Parts:[]articlemodel.TextPart{{Text:"Linked article",URL:"https://example.org"}}},{Kind:"img",Role:"editorial",URL:"https://example.org/scene.jpg"}}}
 contributors.SeparateHeader(&a)
 if len(a.Blocks)!=3 {t.Fatalf("unexpected removal %+v",a.Blocks)}
}

const fxnprcontributorpartsnprWestBankInline = "https://npr.brightspotcdn.com/dims3/default/strip/false/crop/4000x3000+0+0/resize/1100/quality/50/format/jpeg/?url=http%3A%2F%2Fnpr-brightspot.s3.amazonaws.com%2Fb9%2Fe3%2F4f3b526d4bbeb6ccd150d42f81c5%2F20261021-102633.jpg"
const fxnprcontributorpartsnprWestBankFirst = "https://npr.brightspotcdn.com/dims3/default/strip/false/crop/836x470+0+78/resize/1400/quality/85/format/jpeg/?url=http%3A%2F%2Fnpr-brightspot.s3.amazonaws.com%2F3d%2F13%2F662de40b41cda580069a58be1bbf%2Fimg-2213.jpg"
const fxnprcontributorpartsestrinPortrait = "https://media.npr.org/assets/img/2020/01/02/daniel-estrin_sq-4199304dee8cdf91935a65a2d0b069e07311386e.jpg?s=100&c=85&f=jpeg"
