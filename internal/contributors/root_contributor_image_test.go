package contributors_test

import (
 articlemodel "clean-article-reader/internal/article"
 "testing"
 "clean-article-reader/internal/contributors"
 "clean-article-reader/internal/documentmedia"
 "clean-article-reader/internal/media"
)


func TestContributorPortraitNotEditorialLead(t *testing.T) {
 a:=articlemodel.Article{ImageURL:fxcontributorimagenprWestBankMetadata,Blocks:[]articlemodel.Block{
  {Kind:"img",Role:"author",URL:"https://example.org/samantha.png",Text:"Photo of Samantha Balaban in the field"},
  {Kind:"p",Text:"Daniel Estrin",Links:[]articlemodel.InlineLink{{Text:"Daniel Estrin",URL:"https://www.npr.org/people/123"}}},
  {Kind:"img",Text:"Daniel Estrin",URL:fxcontributorimageestrinPortrait},
  {Kind:"p",Text:"Daniel Estrin"},
  {Kind:"img",Role:"lead",URL:fxcontributorimagenprWestBankFirst,Text:"Israeli security forces"},
  {Kind:"p",Text:"Story reporting"},
  {Kind:"img",URL:fxcontributorimagenprWestBankInline,Text:"Activists and rabbis from the United States and Israel help Palestinians harvest olives"},
 }}
 contributors.ClassifyContributorImages(&a)
 if a.Blocks[2].Role!="author"{t.Fatalf("contributor portrait not classified: %+v",a.Blocks[2])}
 documentmedia.NormalizeDocument(&a,media.SourceIdentity,fxcontributorimagenprWestBankMetadata)
 if a.ImageURL!=fxcontributorimagenprWestBankFirst {t.Fatalf("wrong article lead %s",a.ImageURL)}
 if a.Blocks[2].Role!="author" || a.Blocks[4].Role!="lead" || a.Blocks[6].URL!=fxcontributorimagenprWestBankInline {t.Fatalf("image roles or order wrong: %+v",a.Blocks)}
}
func TestEditorialCaptionNotPromotedToContributor(t *testing.T) {
 a:=articlemodel.Article{Blocks:[]articlemodel.Block{{Kind:"p",Text:"Photo by Daniel Estrin"},{Kind:"img",Text:"Military patrols in the area",URL:fxcontributorimagenprWestBankFirst}}}
 contributors.ClassifyContributorImages(&a)
 if a.Blocks[1].Role=="author"{t.Fatal("editorial figure incorrectly classified")}
}

const fxcontributorimagenprWestBankMetadata = "https://npr.brightspotcdn.com/dims3/default/strip/false/crop/4000x2250+0+375/resize/1400/quality/85/format/jpeg/?url=http%3A%2F%2Fnpr-brightspot.s3.amazonaws.com%2Fb9%2Fe3%2F4f3b526d4bbeb6ccd150d42f81c5%2F20261021-102633.jpg"
const fxcontributorimagenprWestBankInline = "https://npr.brightspotcdn.com/dims3/default/strip/false/crop/4000x3000+0+0/resize/1100/quality/50/format/jpeg/?url=http%3A%2F%2Fnpr-brightspot.s3.amazonaws.com%2Fb9%2Fe3%2F4f3b526d4bbeb6ccd150d42f81c5%2F20261021-102633.jpg"
const fxcontributorimagenprWestBankFirst = "https://npr.brightspotcdn.com/dims3/default/strip/false/crop/836x470+0+78/resize/1400/quality/85/format/jpeg/?url=http%3A%2F%2Fnpr-brightspot.s3.amazonaws.com%2F3d%2F13%2F662de40b41cda580069a58be1bbf%2Fimg-2213.jpg"
const fxcontributorimageestrinPortrait = "https://media.npr.org/assets/img/2020/01/02/daniel-estrin_sq-4199304dee8cdf91935a65a2d0b069e07311386e.jpg?s=100&c=85&f=jpeg"
