package recovery_test

import (
 articlemodel "clean-article-reader/internal/article"
 "clean-article-reader/internal/documentmedia"
 "clean-article-reader/internal/media"
 "testing"
 "clean-article-reader/internal/recovery"
)

func TestRealNPRSourceContainerRoles(t *testing.T) {
 raw:=`<!doctype html><html><head><title>West Bank report</title><meta property="og:image" content="`+fxsourceimagerolesnprWestBankFirst+`"></head><body><main><article><section class="byline"><a href="/people/123"><picture><img src="`+fxsourceimagerolesestrinPortrait+`" alt="Daniel Estrin"></picture></a></section><div class="imagewrap has-source-dimensions"><picture><img src="`+fxsourceimagerolesnprWestBankFirst+`" alt="Israeli security forces"></picture></div><p>Article paragraph.</p><div class="imagewrap has-source-dimensions"><picture><img src="`+fxsourceimagerolesnprWestBankInline+`" alt="Olive harvesting"></picture></div></article></main></body></html>`
 a:=articlemodel.Article{ImageURL:fxsourceimagerolesnprWestBankMetadata,Blocks:[]articlemodel.Block{{Kind:"img",URL:fxsourceimagerolesestrinPortrait,Text:"Daniel Estrin"},{Kind:"img",URL:fxsourceimagerolesnprWestBankFirst,Text:"Israeli security forces"},{Kind:"p",Text:"Article paragraph."},{Kind:"img",URL:fxsourceimagerolesnprWestBankInline,Text:"Olive harvesting"}}}
 recovery.ClassifyImagesFromSource(&a,raw,"https://www.npr.org/article")
 if a.Blocks[0].Role!="author" || a.Blocks[1].Role!="editorial" || a.Blocks[3].Role!="editorial" {t.Fatalf("source roles wrong: %+v",a.Blocks)}
 documentmedia.NormalizeDocument(&a,media.SourceIdentity,fxsourceimagerolesnprWestBankMetadata)
 if a.ImageURL!=fxsourceimagerolesnprWestBankFirst || a.Blocks[0].Role!="author" || a.Blocks[1].Role!="lead" || a.Blocks[3].Role!="editorial" {t.Fatalf("incorrect image precedence: %+v",a.Blocks)}
}
func TestUnclassifiedContributorCannotOverrideConfirmedEditorial(t *testing.T) {
 a:=articlemodel.Article{ImageURL:fxsourceimagerolesnprWestBankMetadata,Blocks:[]articlemodel.Block{{Kind:"img",URL:fxsourceimagerolesestrinPortrait,Text:"Daniel Estrin"},{Kind:"img",Role:"editorial",URL:fxsourceimagerolesnprWestBankFirst,Text:"Israeli security forces"}}}
 documentmedia.NormalizeDocument(&a,media.SourceIdentity,fxsourceimagerolesnprWestBankMetadata)
 if a.ImageURL!=fxsourceimagerolesnprWestBankFirst||a.Blocks[1].Role!="lead" {t.Fatalf("portrait displaced source-confirmed figure: %+v",a.Blocks)}
}

const fxsourceimagerolesnprWestBankMetadata = "https://npr.brightspotcdn.com/dims3/default/strip/false/crop/4000x2250+0+375/resize/1400/quality/85/format/jpeg/?url=http%3A%2F%2Fnpr-brightspot.s3.amazonaws.com%2Fb9%2Fe3%2F4f3b526d4bbeb6ccd150d42f81c5%2F20261021-102633.jpg"
const fxsourceimagerolesnprWestBankInline = "https://npr.brightspotcdn.com/dims3/default/strip/false/crop/4000x3000+0+0/resize/1100/quality/50/format/jpeg/?url=http%3A%2F%2Fnpr-brightspot.s3.amazonaws.com%2Fb9%2Fe3%2F4f3b526d4bbeb6ccd150d42f81c5%2F20261021-102633.jpg"
const fxsourceimagerolesnprWestBankFirst = "https://npr.brightspotcdn.com/dims3/default/strip/false/crop/836x470+0+78/resize/1400/quality/85/format/jpeg/?url=http%3A%2F%2Fnpr-brightspot.s3.amazonaws.com%2F3d%2F13%2F662de40b41cda580069a58be1bbf%2Fimg-2213.jpg"
const fxsourceimagerolesestrinPortrait = "https://media.npr.org/assets/img/2020/01/02/daniel-estrin_sq-4199304dee8cdf91935a65a2d0b069e07311386e.jpg?s=100&c=85&f=jpeg"
