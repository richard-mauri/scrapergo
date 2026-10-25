package reader

import (
 "reflect"
 "testing"

 "clean-article-reader/internal/article"
 "clean-article-reader/internal/extraction"
)

func TestProcessorStageOrderAndEvidence(t *testing.T) {
 var events []string
 src := func(name string) SourceTransform {
  return func(a *article.Article, raw, url string) {
   if raw != "<article>fixture</article>" || url != "https://example.org/article" {
    t.Fatalf("source evidence lost for %s: raw=%q url=%q",name,raw,url)
   }
   events=append(events,name)
  }
 }
 plain := func(name string) func(*article.Article) { return func(*article.Article) { events=append(events,name) } }
 p:=New(Dependencies{
  RecoverPhotographs:src("recover-photographs"),
  RecoverLinks:src("recover-links"),
  RecoverAuthorPortrait:src("recover-author-portrait"),
  ClassifyContributors:plain("classify-contributors"),
  ClassifySourceImages:src("classify-source-images"),
  ReconcileByline:plain("reconcile-byline"),
  RecoverOpeningFigure:src("recover-opening-figure"),
  Normalize:func(_ *article.Article, _ string){events=append(events,"normalize")},
  RestoreGalleryGroups:src("restore-gallery-groups"),
 })
 want:=[]string{
  "recover-photographs","deduplicate-images","recover-links","recover-author-portrait",
  "remove-player-blocks","remove-access-chrome","classify-contributors","classify-source-images",
  "exclude-poststory-author-portraits","reconcile-byline","separate-contributor-header",
  "recover-opening-figure","normalize","restore-gallery-groups",
 }
 stages:=p.Stages()
 got:=make([]string,len(stages))
 for i,s:=range stages {got[i]=s.Name()}
 if !reflect.DeepEqual(got,want) {t.Fatalf("stage order changed: got %v want %v",got,want)}
 result:=extraction.Result{Article:article.Article{URL:"https://example.org/article",Blocks:[]article.Block{{Kind:"p",Text:"preserved"}}},Evidence:extraction.SourceEvidence{SourceURL:"https://example.org/article"}}
 out:=p.Process(result,"<article>fixture</article>")
 if len(out.Blocks)==0 || out.Blocks[0].Text!="preserved" { t.Fatalf("article lost in pipeline: %+v",out.Blocks) }
 wantEvents:=[]string{"recover-photographs","recover-links","recover-author-portrait","classify-contributors","classify-source-images","reconcile-byline","recover-opening-figure","normalize","restore-gallery-groups"}
 if !reflect.DeepEqual(events,wantEvents){t.Fatalf("callback order changed: got %v want %v",events,wantEvents)}
}
