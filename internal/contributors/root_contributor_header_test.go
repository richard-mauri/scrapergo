package contributors_test

import (
 articlemodel "clean-article-reader/internal/article"
 "testing"
 "strings"
 "clean-article-reader/internal/contributors"
)

func TestNPRHeaderSeparationPreservesEditorialImages(t *testing.T) {
 linked:=func(name string) articlemodel.Block{return articlemodel.Block{Kind:"p",Text:name,Links:[]articlemodel.InlineLink{{Text:name,URL:"https://www.npr.org/people/123"}}}}
 a:=articlemodel.Article{Byline:"Ruth Sherlock",Blocks:[]articlemodel.Block{
  {Kind:"p",Text:"Middle East"}, {Kind:"p",Text:"By"},
  linked("Ruth Sherlock"),{Kind:"p",Text:","},
  {Kind:"img",Role:"author",URL:"https://example.org/samantha.jpg",Text:"Photo of Samantha Balaban in the field"},
  linked("Samantha Balaban"),{Kind:"p",Text:","},linked("Nuha Musleh"),{Kind:"p",Text:","},
  {Kind:"img",Role:"author",URL:fxcontributorheaderestrinPortrait,Text:"Daniel Estrin"},
  linked("Daniel Estrin"),
  {Kind:"h3",Text:"Brad Lander West Bank"},
  {Kind:"p",Text:"Download"},
  {Kind:"img",Role:"editorial",URL:fxcontributorheadernprWestBankFirst,Text:"Israeli security forces"},
  {Kind:"p",Text:"The actual reporting begins here."},
  {Kind:"img",Role:"editorial",URL:fxcontributorheadernprWestBankInline,Text:"Olive harvesting"},
 }}
 contributors.SeparateHeader(&a)
 if a.Byline!="Ruth Sherlock, Samantha Balaban, Nuha Musleh, Daniel Estrin" {t.Fatalf("byline=%q",a.Byline)}
 for _,b:=range a.Blocks {if b.Role=="author"||b.Text=="By"||b.Text==","||b.Text=="Daniel Estrin"{t.Fatalf("header fragment remains: %+v",b)}}
 editorial:=0
 for _,b:=range a.Blocks {if b.Role=="editorial"{editorial++}}
 if editorial!=2 {t.Fatalf("lost original editorial images: %+v",a.Blocks)}
 if a.Blocks[len(a.Blocks)-1].URL!=fxcontributorheadernprWestBankInline{t.Fatal("later editorial image moved")}
}
func TestNoConfirmedEditorialBoundaryLeavesContentUntouched(t *testing.T){
 a:=articlemodel.Article{Blocks:[]articlemodel.Block{{Kind:"p",Text:"By"},{Kind:"p",Text:"Alice"},{Kind:"p",Text:strings.Repeat("Real article content. ",10)}}}
 contributors.SeparateHeader(&a)
 if len(a.Blocks)!=3{t.Fatalf("no source boundary should leave content intact %+v",a.Blocks)}
}

const fxcontributorheadernprWestBankInline = "https://npr.brightspotcdn.com/dims3/default/strip/false/crop/4000x3000+0+0/resize/1100/quality/50/format/jpeg/?url=http%3A%2F%2Fnpr-brightspot.s3.amazonaws.com%2Fb9%2Fe3%2F4f3b526d4bbeb6ccd150d42f81c5%2F20261021-102633.jpg"
const fxcontributorheadernprWestBankFirst = "https://npr.brightspotcdn.com/dims3/default/strip/false/crop/836x470+0+78/resize/1400/quality/85/format/jpeg/?url=http%3A%2F%2Fnpr-brightspot.s3.amazonaws.com%2F3d%2F13%2F662de40b41cda580069a58be1bbf%2Fimg-2213.jpg"
const fxcontributorheaderestrinPortrait = "https://media.npr.org/assets/img/2020/01/02/daniel-estrin_sq-4199304dee8cdf91935a65a2d0b069e07311386e.jpg?s=100&c=85&f=jpeg"
