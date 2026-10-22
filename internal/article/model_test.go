package article

import (
 "encoding/json"
 "strings"
 "testing"
)

func TestArticleJSONContract(t *testing.T) {
 a:=Article{Title:"Story",URL:"https://example.org/story",Blocks:[]Block{{Kind:"p",Text:"Hello",Parts:[]TextPart{{Text:"Hello",URL:"https://example.org/source"}}},{Kind:"img",URL:"https://example.org/photo",Role:"lead"}},WordCount:1}
 payload,err:=json.Marshal(a)
 if err!=nil {t.Fatal(err)}
 for _,term:=range []string{`"title":"Story"`,`"kind":"img"`,`"role":"lead"`,`"wordCount":1`} {if !strings.Contains(string(payload),term){t.Fatalf("missing %s in %s",term,payload)}}
 var got Article
 if err=json.Unmarshal(payload,&got);err!=nil{t.Fatal(err)}
 if got.Blocks[1].Role!="lead" || got.Blocks[0].Parts[0].URL!="https://example.org/source" {t.Fatalf("roundtrip lost data: %+v",got)}
}
