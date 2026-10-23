package main

import (
 articlemodel "clean-article-reader/internal/article"
 "clean-article-reader/internal/recovery"
 "strings"
 "testing"
 "clean-article-reader/internal/media"
)

func TestAuthorBioPortraitBelowEditorialLeadExcluded(t *testing.T) {
 const source = `<html><body><article><figure class="imagewrap"><img src="/images/lead.jpg" alt="Calley Means"/></figure><p>Editorial reporting paragraph with enough words to establish the story prose as the primary body and not a contributor header.</p><section class="author-bio"><a href="/author/beth/"><img src="/images/beth.jpg" alt="Beth Mole"/></a><p>Beth Mole Senior Health Reporter</p></section></article></body></html>`
 a:=articlemodel.Article{Byline:"Beth Mole",Blocks:[]articlemodel.Block{{Kind:"img",URL:"https://example.org/images/lead.jpg",Role:"lead",Text:"Calley Means"},{Kind:"p",Text:strings.Repeat("Story body information. ",7)},{Kind:"img",URL:"https://example.org/images/beth.jpg",Text:"Beth Mole"}}}
 recovery.ClassifyImagesFromSource(&a,source,"https://example.org/report")
 if a.Blocks[2].Role!="author" {t.Fatalf("bio photo not classified: %+v",a.Blocks)}
 media.RemovePostEditorialAuthorPortraits(&a)
 if len(a.Blocks)!=2 || a.Blocks[0].Role!="lead" || a.Byline!="Beth Mole" {t.Fatalf("editorial lead or byline changed: %+v",a)}
}

func TestNPRContributorPortraitBeforeLeadPreserved(t *testing.T) {
 a:=articlemodel.Article{Byline:"Daniel Estrin",Blocks:[]articlemodel.Block{{Kind:"img",Role:"author",URL:estrinPortrait,Text:"Daniel Estrin"},{Kind:"img",Role:"lead",URL:nprWestBankFirst},{Kind:"p",Text:"Story body"}}}
 media.RemovePostEditorialAuthorPortraits(&a)
 if len(a.Blocks)!=3 || a.Blocks[0].Role!="author" || a.Blocks[1].Role!="lead" {t.Fatalf("NPR contributor lost: %+v",a.Blocks)}
}

func TestEditorialFiguresAfterLeadUnaffected(t *testing.T) {
 a:=articlemodel.Article{Blocks:[]articlemodel.Block{{Kind:"img",Role:"lead",URL:"https://example.org/lead.jpg"},{Kind:"img",Role:"editorial",URL:"https://example.org/inside.jpg"},{Kind:"p",Text:"Content"}}}
 media.RemovePostEditorialAuthorPortraits(&a)
 if len(a.Blocks)!=3 {t.Fatalf("editorial figure removed: %+v",a.Blocks)}
}
