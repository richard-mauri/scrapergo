package main

import (
 articlemodel "clean-article-reader/internal/article"
 "clean-article-reader/internal/documentmedia"
 "clean-article-reader/internal/media"
 "strings"
 "testing"
)

const nprLead = "https://npr.brightspotcdn.com/dims3/default/strip/false/crop/8640x4860+0+450/resize/1400/quality/85/format/jpeg/?url=http%3A%2F%2Fnpr-brightspot.s3.amazonaws.com%2Fdb%2F63%2Fd779a1c14ad9885a895acb202701%2Fgettyimages-2285931973.jpg"
const nprInline = "https://npr.brightspotcdn.com/dims3/default/strip/false/crop/8640x5760+0+0/resize/1100/quality/50/format/jpeg/?url=http%3A%2F%2Fnpr-brightspot.s3.amazonaws.com%2Fdb%2F63%2Fd779a1c14ad9885a895acb202701%2Fgettyimages-2285931973.jpg"
const nprAuthor = "https://npr.brightspotcdn.com/dims3/default/strip/false/crop/2673x2673+254+0/resize/100/quality/85/format/jpeg/?url=http%3A%2F%2Fnpr-brightspot.s3.amazonaws.com%2F0f%2Fc7%2Fa5aefd694b64955eeec162b7840f%2Fstephen-fowler.JPG"

func TestNPRSourceImageIdentity(t *testing.T) {
 if media.SourceIdentity(nprLead)!=media.SourceIdentity(nprInline) {t.Fatal("NPR variants were not matched by actual source asset")}
 if media.SourceIdentity(nprAuthor)==media.SourceIdentity(nprLead) {t.Fatal("author portrait incorrectly matched lead")}
}
func TestNPRLeadCaptionAndPortrait(t *testing.T) {
 caption:="In this file photo from July 2026, President Trump shows off an executive order he signed in the Oval Office of the White House."
 a:=articlemodel.Article{Title:"A report",ImageURL:nprLead,Blocks:[]articlemodel.Block{
  {Kind:"img",Role:"author",URL:nprAuthor,Text:"Headshot of Stephen Fowler"},
  {Kind:"img",URL:nprInline,Text:caption},
  {Kind:"p",Text:"Reporting continues."},
 }}
 documentmedia.NormalizeDocument(&a,media.SourceIdentity,nprLead)
 images,leads,authors:=0,0,0
 for _,b:=range a.Blocks {
  if b.Kind!="img"{continue}
  images++
  if b.Role=="lead" {leads++;if b.URL!=nprLead||b.Text!=caption {t.Fatalf("lead rendition/caption mismatch: %+v",b)}}
  if b.Role=="author" {authors++;if !strings.Contains(b.URL,"stephen-fowler.JPG"){t.Fatalf("wrong author: %+v",b)}}
 }
 if images!=2||leads!=1||authors!=1 {t.Fatalf("images=%d leads=%d authors=%d blocks=%+v",images,leads,authors,a.Blocks)}
}
func TestUnrelatedImagesRemainDistinct(t *testing.T) {
 a:=articlemodel.Article{Blocks:[]articlemodel.Block{{Kind:"img",URL:"https://example.com/a.jpg"},{Kind:"img",URL:"https://example.com/b.jpg"}}}
 documentmedia.NormalizeDocument(&a,media.SourceIdentity)
 if len(a.Blocks)!=2{t.Fatalf("distinct photographs lost: %+v",a.Blocks)}
}
