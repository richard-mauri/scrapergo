package bootstrap_test

import (
 articlemodel "clean-article-reader/internal/article"
 "clean-article-reader/internal/bootstrap"
 "encoding/json"
 "net/http/httptest"
 "strings"
 "testing"
)

func TestStructuredPipelinePreservesArticleWithoutEmbeddedMedia(t *testing.T) {
 prose:=strings.Repeat("The report describes changing conditions and includes context from multiple sources. ",12)
 source:=`<!doctype html><html><head><title>Example reporting headline</title><meta property="og:title" content="Example reporting headline"><meta property="og:image" content="https://news.example/lead.jpg"></head><body><main><article><h1>Example reporting headline</h1><p>`+prose+`</p><div><iframe src="https://audio.example/player"></iframe></div><p>Subsequent details and independent analysis are preserved in full. `+prose+`</p></article></main></body></html>`
 rec:=httptest.NewRecorder()
 rec.Header().Set("Content-Type","application/json")
 bootstrap.WriteExtractedArticle(rec,source,"https://news.example/report")
 if rec.Code!=200 {t.Fatalf("response=%d %s",rec.Code,rec.Body.String())}
 var article articlemodel.Article
 if err:=json.Unmarshal(rec.Body.Bytes(),&article);err!=nil{t.Fatal(err)}
 for _,b:=range article.Blocks {
  if strings.Contains(b.Text,"<iframe")||strings.Contains(b.Text,"audio.example/player"){t.Fatalf("media leaked: %q",b.Text)}
 }
 joined:=""
 for _,b:=range article.Blocks {joined+=b.Text+" "}
 for _,needle:=range []string{"The report describes changing conditions","Subsequent details and independent analysis"} {
  if !strings.Contains(joined,needle){t.Fatalf("lost editorial text %q",needle)}
 }
}
