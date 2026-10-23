package main

import (
 articlemodel "clean-article-reader/internal/article"
 "clean-article-reader/internal/bootstrap"
 "encoding/json"
 "net/http/httptest"
 "os"
 "strings"
 "testing"
)

func TestFixtureArticleMediaAndAuthor(t *testing.T){
 raw,err:=os.ReadFile("testdata/article_media_and_author.html")
 if err!=nil{t.Fatal(err)}
 rec:=httptest.NewRecorder()
 bootstrap.WriteExtractedArticle(rec,string(raw),"https://news.example.org/report")
 if rec.Code!=200{t.Fatalf("HTTP %d: %s",rec.Code,rec.Body.String())}
 var a articlemodel.Article
 if err:=json.Unmarshal(rec.Body.Bytes(),&a);err!=nil{t.Fatal(err)}
 leadCount,authorCount:=0,0
 for _,b:=range a.Blocks {
  if strings.Contains(b.Text,"audio.example.org/player")||strings.Contains(b.Text,"<iframe"){t.Fatalf("embedded media markup leaked: %+v",b)}
  if b.Kind=="img"&&strings.Contains(b.URL,"lead-courthouse.jpg"){leadCount++;if b.Role!="lead"{t.Errorf("lead image missing role: %+v",b)}}
  if b.Kind=="img"&&strings.Contains(b.URL,"author-portrait.jpg"){authorCount++;if b.Role!="author"{t.Errorf("author image missing role: %+v",b)}}
 }
 if leadCount!=1{t.Errorf("lead count = %d, want exactly 1",leadCount)}
 if authorCount!=1{t.Errorf("author portrait count = %d, want exactly 1",authorCount)}
 if a.WordCount<60{t.Errorf("article text too short: %d",a.WordCount)}
}
