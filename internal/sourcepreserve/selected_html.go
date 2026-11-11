// Package sourcepreserve also houses the offline-only selected-HTML experiment.
// This structural document is NOT the production article model or safe HTML.
package sourcepreserve

import (
 "fmt"
 "net/url"
 "strings"

 "clean-article-reader/internal/sourcehtml"
 "golang.org/x/net/html"
)

// SemanticNode preserves only explicitly present DOM semantics. Children retain
// order and ownership; it contains no publisher HTML, CSS, handlers or scripts.
type SemanticNode struct {
 Kind string `json:"kind"`
 Text string `json:"text,omitempty"`
 URL string `json:"url,omitempty"`
 DateTime string `json:"dateTime,omitempty"`
 Children []SemanticNode `json:"children,omitempty"`
}
type SemanticDocument struct {
 Nodes []SemanticNode `json:"nodes"`
 UnsupportedImages int `json:"unsupportedImages,omitempty"`
 UnsafeURLs int `json:"unsafeURLs,omitempty"`
}

// ConvertSelectedHTML converts Readability's RenderHTML output, NOT source
// HTML selection. It is an isolated research entry point; it never fetches,
// renders, mutates canonical blocks or guesses a missing caption association.
func ConvertSelectedHTML(selectedHTML,baseURL string)(SemanticDocument,error){
 var out SemanticDocument
 u,err:=url.Parse(baseURL)
 if err!=nil || (u.Scheme!="http" && u.Scheme!="https") || u.Hostname()=="" || u.User!=nil {
  return out,fmt.Errorf("invalid HTTP(S) base URL")
 }
 root,err:=html.Parse(strings.NewReader(selectedHTML))
 if err!=nil {return out,err}
 body:=find(root,func(n *html.Node)bool{return element(n,"body")})
 if len(body)==0 {return out,fmt.Errorf("missing HTML body")}
 // Parser's synthetic body is the selected DOM root; never select a
 // publisher article here or infer hidden content outside Readability output.
 for child:=body[0].FirstChild;child!=nil;child=child.NextSibling {
  out.Nodes=append(out.Nodes,convertNode(child,u,&out)...)
 }
 return out,nil
}

func convertNode(n *html.Node,base *url.URL,report *SemanticDocument)[]SemanticNode{
 if n==nil||excluded(n){return nil}
 if n.Type==html.TextNode {
  text:=sourcehtml.Normalize(n.Data)
  if text==""{return nil}
  return []SemanticNode{{Kind:"text",Text:text}}
 }
 if n.Type!=html.ElementNode{return nil}
 tag:=n.Data
 switch tag {
 case "img":
  url:=sourcehtml.ResolveURL(attr(n,"src"),base)
  if url=="" {report.UnsafeURLs++;return nil}
  // A bare image is retained without a claimed caption or figure owner.
  report.UnsupportedImages++
  return []SemanticNode{{Kind:"image",URL:url}}
 case "script","style","iframe","form","button","object","embed","svg":return nil
 }
 // Containers, headings, paragraphs, inline semantic links/formatting,
 // lists/tables and explicit figures are retained as a tree. Unknown wrappers
 // are transparent; no CSS/class-based significance is inferred.
 kind:=map[string]string{
  "article":"article","main":"main","section":"section",
  "p":"paragraph","h1":"h1","h2":"h2","h3":"h3",
  "h4":"h4","h5":"h5","h6":"h6","blockquote":"blockquote",
  "ul":"unordered-list","ol":"ordered-list","li":"list-item",
  "table":"table","thead":"table-head","tbody":"table-body","tfoot":"table-foot",
  "tr":"table-row","th":"table-header","td":"table-cell",
  "figure":"figure","figcaption":"figcaption","time":"time",
  "a":"link","strong":"strong","b":"strong","em":"emphasis","i":"emphasis",
  "code":"code","pre":"pre","br":"line-break",
  "small":"small","cite":"cite",
 }[tag]
 if kind=="" {
  var children []SemanticNode
  for c:=n.FirstChild;c!=nil;c=c.NextSibling{children=append(children,convertNode(c,base,report)...)}
  return children
 }
 node:=SemanticNode{Kind:kind}
 if tag=="a" {
  node.URL=sourcehtml.ResolveURL(attr(n,"href"),base)
  if attr(n,"href")!="" && node.URL=="" {report.UnsafeURLs++;node.Kind="span"}
 }
 if tag=="time" {node.DateTime=attr(n,"datetime")}
 for c:=n.FirstChild;c!=nil;c=c.NextSibling{node.Children=append(node.Children,convertNode(c,base,report)...)}
 // Structural elements without children (other than a real line break) are
 // not editorial content. No caption is fabricated for empty figures.
 if len(node.Children)==0 && kind!="line-break" {return nil}
 return []SemanticNode{node}
}
