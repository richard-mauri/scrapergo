package contributors_test

import (
 articlemodel "clean-article-reader/internal/article"
 "clean-article-reader/internal/contributors"
 "clean-article-reader/internal/reader"
 "strings"
 "testing"
 "clean-article-reader/internal/extraction"
)

func TestRecoverAuthorPortraitRejectsPoststoryBiography(t *testing.T) {
 raw:=`<html><body><main><article><figure><img src="/lead.jpg" alt="Editorial story photo"/></figure><p>`+strings.Repeat("This editorial paragraph establishes substantive source content. ",4)+`</p><section class="author-bio"><a href="/author/beth"><img src="/beth.jpg" alt="Photo of Beth Mole"/></a></section></article></main></body></html>`
 a:=articlemodel.Article{Byline:"Beth Mole",Blocks:[]articlemodel.Block{{Kind:"p",Text:"Article narrative retained."}}}
 contributors.RecoverAuthorPortrait(&a,raw,"https://example.org/health")
 for _,b:=range a.Blocks {if b.Kind=="img" && strings.Contains(b.URL,"beth.jpg") {t.Fatalf("post-story biography must not be prepended: %+v",a.Blocks)}}
}

func TestRecoverAuthorPortraitRetainsOpeningContributor(t *testing.T) {
 raw:=`<html><body><article><section class="byline"><a href="/people/123"><img src="/estrin.jpg" alt="Daniel Estrin"/></a></section><figure><img src="/story.jpg" alt="Story image"/></figure><p>`+strings.Repeat("Substantive article story prose. ",12)+`</p></article></body></html>`
 a:=articlemodel.Article{Byline:"Daniel Estrin",Blocks:[]articlemodel.Block{{Kind:"p",Text:"Story prose."}}}
 contributors.RecoverAuthorPortrait(&a,raw,"https://example.org/news")
 if len(a.Blocks)!=2 || a.Blocks[0].Role!="author" || a.Blocks[0].URL!="https://example.org/estrin.jpg" {t.Fatalf("opening contributor portrait lost: %+v",a.Blocks)}
}

func TestAssembledPoststoryPortraitAbsentAndLeadPreserved(t *testing.T) {
 raw:=`<html><body><main><article><figure><img src="/lead.jpg" alt="Editorial photograph"/></figure><p>`+strings.Repeat("This article covers an important current development in science. ",7)+`</p><section class="author-bio"><a href="/author/beth"><img src="/beth.jpg" alt="Photo of Beth Mole"/></a></section></article></main></body></html>`
 extracted:=extraction.Result{}
 extracted.Article=articlemodel.Article{Byline:"Beth Mole",Blocks:[]articlemodel.Block{{Kind:"img",URL:"https://example.org/lead.jpg",Role:"lead"},{Kind:"p",Text:strings.Repeat("Editorial story text. ",8)}}}
 extracted.Evidence.SourceURL="https://example.org/health"
 a:=reader.DefaultProcessor().Process(extracted,raw)
 for _,b:=range a.Blocks {if b.Kind=="img" && strings.Contains(b.URL,"beth.jpg") {t.Fatalf("author portrait leaked into final body: %+v",a.Blocks)}}
 if a.Byline!="Beth Mole" || a.ImageURL!="https://example.org/lead.jpg" {t.Fatalf("byline or lead changed: %+v",a)}
}


