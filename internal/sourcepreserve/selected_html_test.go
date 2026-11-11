package sourcepreserve

import (
 "encoding/json"
 "strings"
 "testing"
)

func nodeKinds(nodes []SemanticNode)[]string{
 var kinds []string
 var walk func([]SemanticNode)
 walk=func(items []SemanticNode){
  for _,n:=range items{kinds=append(kinds,n.Kind);walk(n.Children)}
 }
 walk(nodes)
 return kinds
}
func findSemantic(nodes []SemanticNode,kind string)[]SemanticNode{
 var result []SemanticNode
 var walk func([]SemanticNode)
 walk=func(items []SemanticNode){for _,n:=range items{if n.Kind==kind{result=append(result,n)};walk(n.Children)}}
 walk(nodes)
 return result
}
func TestSelectedHTMLKeepsExplicitFigureOwnershipAndInlineLinks(t *testing.T){
 raw:=`<article><h2>Section</h2><p>Introduction <a href="/reference"><em>with link</em></a> and <strong>emphasis</strong>.</p><figure><img src="/a.jpg"><figcaption>Left image <cite>Photographer A</cite></figcaption></figure><figure><img src="/b.jpg"><figcaption>Right image <cite>Photographer B</cite></figcaption></figure></article>`
 got,err:=ConvertSelectedHTML(raw,"https://example.org/story")
 if err!=nil{t.Fatal(err)}
 if len(got.Nodes)!=1||got.Nodes[0].Kind!="article"{t.Fatalf("article lost: %+v",got)}
 article:=got.Nodes[0]
 if len(article.Children)!=4||article.Children[0].Kind!="h2"||article.Children[1].Kind!="paragraph"||article.Children[2].Kind!="figure"||article.Children[3].Kind!="figure" {t.Fatalf("source order lost: %+v",article.Children)}
 links:=findSemantic(got.Nodes,"link")
 if len(links)!=1||links[0].URL!="https://example.org/reference"||len(findSemantic(links[0].Children,"emphasis"))!=1 {t.Fatalf("inline link/formatting lost: %+v",links)}
 figures:=findSemantic(got.Nodes,"figure")
 if len(figures)!=2||len(findSemantic(figures[0].Children,"image"))!=1||len(findSemantic(figures[0].Children,"figcaption"))!=1||len(findSemantic(figures[1].Children,"image"))!=1 {t.Fatalf("figure ownership lost: %+v",figures)}
 if got.ObservedImages!=2 {t.Fatalf("image observation count wrong: %+v",got)}
}
func TestSelectedHTMLPreservesTablesListsAndTime(t *testing.T){
 raw:=`<section><time datetime="2026-11-11T12:00:00Z">Noon</time><ol><li>First</li><li>Second</li></ol><table><thead><tr><th>Model</th><th>Price</th></tr></thead><tbody><tr><td>A</td><td>100</td></tr></tbody></table></section>`
 got,err:=ConvertSelectedHTML(raw,baseURL)
 if err!=nil{t.Fatal(err)}
 for _,kind:=range []string{"time","ordered-list","list-item","table","table-head","table-row","table-header","table-body","table-cell"}{
  if len(findSemantic(got.Nodes,kind))==0{t.Fatalf("%s semantics lost: %v",kind,nodeKinds(got.Nodes))}
 }
 if times:=findSemantic(got.Nodes,"time");len(times)!=1||times[0].DateTime!="2026-11-11T12:00:00Z"{t.Fatalf("time lost: %+v",times)}
}
func TestSelectedHTMLNeverManufacturesCaptionForAdjacentParagraph(t *testing.T){
 raw:=`<div><img src="/photo.jpg"><p>Independent paragraph.</p></div>`
 got,err:=ConvertSelectedHTML(raw,baseURL)
 if err!=nil{t.Fatal(err)}
 if len(findSemantic(got.Nodes,"image"))!=1||len(findSemantic(got.Nodes,"paragraph"))!=1||len(findSemantic(got.Nodes,"figcaption"))!=0||len(findSemantic(got.Nodes,"figure"))!=0{t.Fatalf("inferred figure or lost content: %+v",got)}
}
func TestSelectedHTMLFiltersActiveMarkupAndUnsafeURLs(t *testing.T){
 raw:=`<article><script>privateJavascript()</script><iframe src="https://example.org/embed"></iframe><p><a href="javascript:alert(1)">Unsafe text</a> and <a href="/safe">Safe link</a></p><img src="data:image/svg+xml,evil"><div hidden><p>Hidden advert</p></div><p>Actual editorial text.</p></article>`
 got,err:=ConvertSelectedHTML(raw,baseURL)
 if err!=nil{t.Fatal(err)}
 data,err:=json.Marshal(got);if err!=nil{t.Fatal(err)}
 for _,secret:=range []string{"privateJavascript","evil","Hidden advert","javascript:","data:image","iframe"}{if strings.Contains(string(data),secret){t.Fatalf("unsafe content leaked: %s",data)}}
 if got.UnsafeURLs!=2||len(findSemantic(got.Nodes,"link"))!=1||len(findSemantic(got.Nodes,"image"))!=0||len(findSemantic(got.Nodes,"paragraph"))!=2{t.Fatalf("incorrect semantic filter: %+v",got)}
}
func TestSelectedHTMLRetainsLegitimateRepetitionAndDoesNotClaimSourceScope(t *testing.T){
 got,err:=ConvertSelectedHTML(`<p>Repeated quote.</p><p>Repeated quote.</p>`,baseURL)
 if err!=nil{t.Fatal(err)}
 if len(got.Nodes)!=2||got.Nodes[0].Kind!="paragraph"||got.Nodes[1].Kind!="paragraph"{t.Fatalf("repeated paragraphs disappeared: %+v",got)}
 if _,err:=ConvertSelectedHTML("<p>Hello</p>","javascript:bad");err==nil {t.Fatal("unsafe base URL accepted")}
}
