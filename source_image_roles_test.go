package main

import (
 "testing"
 "clean-article-reader/internal/recovery"
)

func TestRealNPRSourceContainerRoles(t *testing.T) {
 raw:=`<!doctype html><html><head><title>West Bank report</title><meta property="og:image" content="`+nprWestBankFirst+`"></head><body><main><article><section class="byline"><a href="/people/123"><picture><img src="`+estrinPortrait+`" alt="Daniel Estrin"></picture></a></section><div class="imagewrap has-source-dimensions"><picture><img src="`+nprWestBankFirst+`" alt="Israeli security forces"></picture></div><p>Article paragraph.</p><div class="imagewrap has-source-dimensions"><picture><img src="`+nprWestBankInline+`" alt="Olive harvesting"></picture></div></article></main></body></html>`
 a:=Article{ImageURL:nprWestBankMetadata,Blocks:[]Block{{Kind:"img",URL:estrinPortrait,Text:"Daniel Estrin"},{Kind:"img",URL:nprWestBankFirst,Text:"Israeli security forces"},{Kind:"p",Text:"Article paragraph."},{Kind:"img",URL:nprWestBankInline,Text:"Olive harvesting"}}}
 recovery.ClassifyImagesFromSource(&a,raw,"https://www.npr.org/article")
 if a.Blocks[0].Role!="author" || a.Blocks[1].Role!="editorial" || a.Blocks[3].Role!="editorial" {t.Fatalf("source roles wrong: %+v",a.Blocks)}
 normalizeArticleDocument(&a,nprWestBankMetadata)
 if a.ImageURL!=nprWestBankFirst || a.Blocks[0].Role!="author" || a.Blocks[1].Role!="lead" || a.Blocks[3].Role!="editorial" {t.Fatalf("incorrect image precedence: %+v",a.Blocks)}
}
func TestUnclassifiedContributorCannotOverrideConfirmedEditorial(t *testing.T) {
 a:=Article{ImageURL:nprWestBankMetadata,Blocks:[]Block{{Kind:"img",URL:estrinPortrait,Text:"Daniel Estrin"},{Kind:"img",Role:"editorial",URL:nprWestBankFirst,Text:"Israeli security forces"}}}
 normalizeArticleDocument(&a,nprWestBankMetadata)
 if a.ImageURL!=nprWestBankFirst||a.Blocks[1].Role!="lead" {t.Fatalf("portrait displaced source-confirmed figure: %+v",a.Blocks)}
}
