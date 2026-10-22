package structured

import (
 "net/url"
 "strings"
 "testing"

 "clean-article-reader/internal/sourcehtml"
)

func TestTextBlockPreservesFormattingAndURLPolicy(t *testing.T) {
 root:=sourcehtml.Parse(`<article><p>Before <strong>important</strong> and <a href="/report">source</a> after.</p></article>`)
 nodes:=sourcehtml.Find(root,func(n *sourcehtml.Node)bool{return n.Tag=="p"})
 if len(nodes)!=1 {t.Fatalf("nodes: %d",len(nodes))}
 base,_:=url.Parse("https://news.example.test/story")
 resolve:=func(raw string,b *url.URL) string {u,e:=url.Parse(raw);if e!=nil{return ""};return b.ResolveReference(u).String()}
 result:=TextBlock(nodes[0],base,resolve,func(s string)bool{return strings.HasPrefix(s,"https://")})
 if result.Kind!="p" || result.Text!="Before important and source after." {t.Fatalf("text: %+v",result)}
 bold,linked:=false,false
 for _,part:=range result.Parts {
  if part.Bold&&strings.Contains(part.Text,"important"){bold=true}
  if part.URL=="https://news.example.test/report"&&strings.Contains(part.Text,"source"){linked=true}
 }
 if !bold||!linked {t.Fatalf("formatting missing: %+v",result.Parts)}
}
func TestTextBlockRespectsURLPolicy(t *testing.T) {
 root:=sourcehtml.Parse(`<p>Read <a href="/private">the archive</a> for additional context.</p>`)
 n:=sourcehtml.Find(root,func(n *sourcehtml.Node)bool{return n.Tag=="p"})[0]
 result:=TextBlock(n,nil,func(_ string,_ *url.URL)string{return "https://example.org/private"},func(string)bool{return false})
 for _,part:=range result.Parts {if part.URL!="" {t.Fatalf("blocked URL leaked: %+v",part)}}
}
