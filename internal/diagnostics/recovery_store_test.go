package diagnostics

import (
 "os"
 "path/filepath"
 "strings"
 "testing"

 "clean-article-reader/internal/article"
 "clean-article-reader/internal/extraction"
)

func TestRecoveryRecorderPersistsAcrossSessionsWithoutArticleContent(t *testing.T) {
 dir:=filepath.Join(t.TempDir(),"diagnostics")
 articleURL:="https://example.org/story?private-token=never-store"
 a:=article.Article{Title:"Secret article title",Method:"Go Readability v2",URL:articleURL,Blocks:[]article.Block{{Kind:"p",Text:"Private article text"}},WordCount:3}
 trace:=extraction.NewRecoveryTrace()
 trace.Run("institutional-tail",a.Blocks,func() []article.Block {return a.Blocks})
 first,err:=NewRecoveryRecorder(dir,"Build 000001")
 if err!=nil{t.Fatal(err)}
 if err:=first.Record(articleURL,a,nil,trace);err!=nil{t.Fatal(err)}
 second,err:=NewRecoveryRecorder(dir,"Build 000002")
 if err!=nil{t.Fatal(err)}
 if err:=second.Record(articleURL,a,nil,trace);err!=nil{t.Fatal(err)}
 raw,err:=os.ReadFile(filepath.Join(dir,"recovery.jsonl"))
 if err!=nil{t.Fatal(err)}
 lines:=strings.Split(strings.TrimSpace(string(raw)),"\n")
 if len(lines)!=2{t.Fatalf("want two persisted observations, got %d",len(lines))}
 for _,forbidden:=range []string{"private-token","story?","Secret article title","Private article text"} {
  if strings.Contains(string(raw),forbidden){t.Fatalf("diagnostic leaked %q",forbidden)}
 }
 }
}
