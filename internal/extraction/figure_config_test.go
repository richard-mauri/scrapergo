package extraction

import (
 "fmt"
 "net/url"
 "os"
 "strings"
 "testing"

 "clean-article-reader/internal/article"
)

// Each fixture image is a distinct explicit figure. These tests characterize
// the configured candidate limit, NOT a general editorial-selection policy.
func imageLimitFixture(n int) string {
 var b strings.Builder
 b.WriteString("<article>")
 for i:=1;i<=n;i++ {
  fmt.Fprintf(&b,`<figure><img src="/image-%02d.jpg"><figcaption>Caption for image %d</figcaption></figure>`,i,i)
 }
 b.WriteString("</article>")
 return b.String()
}
func TestFigureSelectionConfigValidation(t *testing.T) {
 for _,v:=range []int{-1,0,MaxArticleImageCandidateLimit+1} {
  if _,err:=NewFigureSelectionConfig(v);err==nil {t.Fatalf("accepted invalid limit %d",v)}
 }
 for _,v:=range []int{MinArticleImageCandidateLimit,DefaultArticleImageCandidateLimit,MaxArticleImageCandidateLimit} {
  got,err:=NewFigureSelectionConfig(v);if err!=nil || got.ArticleImageCandidateLimit()!=v {t.Fatalf("limit %d: %+v %v",v,got,err)}
 }
}
func TestFigureSelectionConfigEnvironment(t *testing.T) {
 t.Run("unset default",func(t *testing.T){
  t.Setenv(ArticleImageCandidateLimitEnv,"12")
  if err:=os.Unsetenv(ArticleImageCandidateLimitEnv);err!=nil{t.Fatal(err)}
  got,err:=FigureSelectionConfigFromEnvironment()
  if err!=nil||got.ArticleImageCandidateLimit()!=DefaultArticleImageCandidateLimit {t.Fatalf("unexpected default value: %v %v",got,err)}
 })
 for _,test:=range []struct{raw string;want int;invalid bool}{
  {"1",1,false},{"13",13,false},{"200",200,false},
  {"",0,true},{"0",0,true},{"-2",0,true},{"201",0,true},{"bogus",0,true},{"1.5",0,true},
 } {
  t.Run(fmt.Sprintf("override_%q",test.raw),func(t *testing.T){
   t.Setenv(ArticleImageCandidateLimitEnv,test.raw)
   got,err:=FigureSelectionConfigFromEnvironment()
   if (err!=nil)!=test.invalid {t.Fatalf("override %q: unexpected error %v",test.raw,err)}
   if !test.invalid&&got.ArticleImageCandidateLimit()!=test.want{t.Fatalf("got %d want %d",got.ArticleImageCandidateLimit(),test.want)}
  })
 }
}
func TestArticleImageCandidateLimitBoundaryAndDefaultCompatibility(t *testing.T) {
 base,_:=url.Parse("https://example.org/article")
 source:=imageLimitFixture(14)
 defaults:=ArticleImages(source,base)
 if len(defaults)!=DefaultArticleImageCandidateLimit {t.Fatalf("default images=%d",len(defaults))}
 for _,test:=range []struct{limit,want int}{{1,1},{11,11},{12,12},{13,13},{14,14},{15,14}}{
  cfg,err:=NewFigureSelectionConfig(test.limit);if err!=nil{t.Fatal(err)}
  got:=ArticleImagesWithConfig(source,base,cfg)
  if len(got)!=test.want{t.Errorf("limit=%d got %d images, want %d",test.limit,len(got),test.want)}
  if got[len(got)-1].Text!=fmt.Sprintf("Caption for image %d",test.want){t.Errorf("limit %d lost explicit caption: %+v",test.limit,got[len(got)-1])}
  // Positioning receives the *same* injected limit.
  blocks:=[]article.Block{{Kind:"p",Text:"An independent paragraph remains unchanged."}}
  positioned:=PositionArticleFiguresWithConfig(blocks,source,base,cfg)
  total:=0
  for _,b:=range positioned{if b.Kind=="img"{total++}}
  if total!=test.want{t.Errorf("positioned limit %d: %d images, want %d",test.limit,total,test.want)}
 }
}
func TestArticleImageLimitDoesNotAdmitInvalidOrDuplicateURLs(t *testing.T) {
 base,_:=url.Parse("https://example.org/article")
 source:=`<article><img src="javascript:evil"><img src="/a.jpg"><img src="/a.jpg"><img src="/b.jpg"><img src="data:image/svg+xml,evil"></article>`
 cfg,err:=NewFigureSelectionConfig(20);if err!=nil{t.Fatal(err)}
 images:=ArticleImagesWithConfig(source,base,cfg)
 if len(images)!=2||images[0].URL!="https://example.org/a.jpg"||images[1].URL!="https://example.org/b.jpg" {t.Fatalf("unsafe/duplicate images selected: %+v",images)}
}
func TestActualReadabilityExtractorRejectsInvalidOverride(t *testing.T) {
 t.Setenv(ArticleImageCandidateLimitEnv,"not-an-integer")
 _,err:=ExtractWithReadability(imageLimitFixture(14),"https://example.org/article",LegacyExtract,InlineTextParts)
 if err==nil||!strings.Contains(err.Error(),ArticleImageCandidateLimitEnv){t.Fatalf("expected configuration error, got %v",err)}
}

func TestProductionReadabilityCandidateLimitOverride(t *testing.T) {
 // A substantial ordinary article with 14 explicit illustrations. This test
 // exercises the production Readability path, not only the image helper.
 raw:="<html><body><article><h1>A long technical field report</h1><p>"+
  strings.Repeat("Researchers analyzed measurements from each instrument to explain the observed results. ",22)+
  "</p>"+imageLimitFixture(14)+"<p>"+
  strings.Repeat("The report continues with details of the analysis and its practical implications. ",22)+
  "</p></article></body></html>"
 read:=func(limit string) int {
  t.Helper()
  t.Setenv(ArticleImageCandidateLimitEnv,limit)
  trace:=NewRecoveryTrace()
  got,err:=ExtractWithReadabilityTraced(raw,"https://example.org/report",LegacyExtract,InlineTextParts,trace)
  if err!=nil{t.Fatal(err)}
  if got.Method!="Go Readability v2"{t.Fatalf("unexpected fallback: %s",got.Method)}
  for _,stage:=range trace.Results(){
   if stage.Stage=="position-figures" {
    if stage.Outcome=="not_reached"{t.Fatalf("position-figures not reached at limit %s",limit)}
    return stage.AfterImages
   }
  }
  t.Fatalf("position-figures trace missing at limit %s",limit)
  return -1
 }
 atDefault:=read("12")
 atHigher:=read("14")
 if atDefault!=12||atHigher!=14{t.Fatalf("position-figures image counts default=%d configured=%d (want 12,14)",atDefault,atHigher)}

}
