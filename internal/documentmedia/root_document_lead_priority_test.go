package documentmedia_test

import (
 articlemodel "clean-article-reader/internal/article"
 "testing"
 "clean-article-reader/internal/documentmedia"
 "clean-article-reader/internal/media"
)


func TestFirstEditorialImageOverridesMetadata(t *testing.T) {
 a:=articlemodel.Article{ImageURL:fxdocumentleadprioritynprWestBankMetadata,Blocks:[]articlemodel.Block{
  {Kind:"img",Role:"author",URL:"https://example.org/author.jpg"},
  {Kind:"p",Text:"Opening article"},
  {Kind:"img",Role:"lead",URL:fxdocumentleadprioritynprWestBankFirst,Text:"Israeli security forces"},
  {Kind:"p",Text:"Further reporting"},
  {Kind:"img",URL:fxdocumentleadprioritynprWestBankInline,Text:"Activists and rabbis harvesting olives"},
 }}
 documentmedia.NormalizeDocument(&a,media.SourceIdentity,fxdocumentleadprioritynprWestBankMetadata)
 if a.ImageURL!=fxdocumentleadprioritynprWestBankFirst {t.Fatalf("metadata displaced editorial lead: %s",a.ImageURL)}
 leadCount,oliveCount:=0,0
 for _,b:=range a.Blocks {
  if b.Kind=="img"&&b.Role=="lead" {leadCount++;if b.URL!=fxdocumentleadprioritynprWestBankFirst {t.Fatalf("wrong lead: %+v",b)}}
  if b.Kind=="img"&&media.SourceIdentity(b.URL)==media.SourceIdentity(fxdocumentleadprioritynprWestBankInline) {oliveCount++;if b.Text!="Activists and rabbis harvesting olives" {t.Fatalf("lost later image caption: %+v",b)}}
 }
 if leadCount!=1||oliveCount!=1||a.Blocks[0].Role!="author" {t.Fatalf("roles/order corrupted: %+v",a.Blocks)}
}
func TestMetadataFallbackWhenNoEditorialImage(t *testing.T) {
 a:=articlemodel.Article{ImageURL:fxdocumentleadprioritynprWestBankMetadata,Blocks:[]articlemodel.Block{{Kind:"img",Role:"author",URL:"https://example.org/author.jpg"},{Kind:"p",Text:"Article"}}}
 documentmedia.NormalizeDocument(&a,media.SourceIdentity,fxdocumentleadprioritynprWestBankMetadata)
 if len(a.Blocks)!=3||a.Blocks[0].Role!="lead"||a.Blocks[1].Role!="author"{t.Fatalf("fallback failed: %+v",a.Blocks)}
}
func TestFirstUnmarkedEditorialImageIsLead(t *testing.T){
 a:=articlemodel.Article{Blocks:[]articlemodel.Block{{Kind:"img",Role:"author",URL:"https://example.org/portrait.jpg"},{Kind:"img",URL:"https://example.org/first.jpg"},{Kind:"img",URL:"https://example.org/second.jpg"}}}
 documentmedia.NormalizeDocument(&a,media.SourceIdentity,fxdocumentleadprioritynprWestBankMetadata)
 if a.ImageURL!="https://example.org/first.jpg"||a.Blocks[1].Role!="lead"||a.Blocks[2].Role=="lead" {t.Fatalf("document order lost: %+v",a)}
}

const fxdocumentleadprioritynprWestBankMetadata = "https://npr.brightspotcdn.com/dims3/default/strip/false/crop/4000x2250+0+375/resize/1400/quality/85/format/jpeg/?url=http%3A%2F%2Fnpr-brightspot.s3.amazonaws.com%2Fb9%2Fe3%2F4f3b526d4bbeb6ccd150d42f81c5%2F20261021-102633.jpg"
const fxdocumentleadprioritynprWestBankInline = "https://npr.brightspotcdn.com/dims3/default/strip/false/crop/4000x3000+0+0/resize/1100/quality/50/format/jpeg/?url=http%3A%2F%2Fnpr-brightspot.s3.amazonaws.com%2Fb9%2Fe3%2F4f3b526d4bbeb6ccd150d42f81c5%2F20261021-102633.jpg"
const fxdocumentleadprioritynprWestBankFirst = "https://npr.brightspotcdn.com/dims3/default/strip/false/crop/836x470+0+78/resize/1400/quality/85/format/jpeg/?url=http%3A%2F%2Fnpr-brightspot.s3.amazonaws.com%2F3d%2F13%2F662de40b41cda580069a58be1bbf%2Fimg-2213.jpg"
