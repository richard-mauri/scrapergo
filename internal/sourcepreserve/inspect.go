// Package sourcepreserve is an offline-only, experimental structural inspector.
//
// It is intentionally not wired to ScraperGo's production extraction or UI.
// It records conservative source-DOM semantics without emitting publisher HTML.
package sourcepreserve

import (
 "fmt"
 "net/url"
 "strings"

 "clean-article-reader/internal/sourcehtml"
 "golang.org/x/net/html"
)

type Item struct {
 Kind string `json:"kind"`
 Text string `json:"text,omitempty"`
 ImageURL string `json:"imageUrl,omitempty"`
 Caption string `json:"caption,omitempty"`
 Credit string `json:"credit,omitempty"`
 Links []string `json:"links,omitempty"`
}

type Report struct {
 Status string `json:"status"`
 Items []Item `json:"items,omitempty"`
 AmbiguousFigures int `json:"ambiguousFigures,omitempty"`
 UnwrappedImages int `json:"unwrappedImages,omitempty"`
 CompactImageTextWrappers int `json:"compactImageTextWrappers,omitempty"`
 OtherUnwrappedImages int `json:"otherUnwrappedImages,omitempty"`
 OtherInMultiImageGroups int `json:"otherInMultiImageGroups,omitempty"`
 OtherWithoutBoundedGroup int `json:"otherWithoutBoundedGroup,omitempty"`
 MultiImageFigures int `json:"multiImageFigures,omitempty"`
 InvalidFigureImages int `json:"invalidFigureImages,omitempty"`
 MultipleCaptions int `json:"multipleCaptions,omitempty"`
 SkippedNoneditorial int `json:"skippedNoneditorial,omitempty"`
}

func element(n *html.Node,tag string) bool {return n!=nil && n.Type==html.ElementNode && n.Data==tag}
func attr(n *html.Node,key string) string {
 if n==nil{return ""}
 for _,a:=range n.Attr {if a.Key==key{return a.Val}}
 return ""
}
func excluded(n *html.Node) bool {
 if n==nil || n.Type!=html.ElementNode{return false}
 switch n.Data {
 case "script","style","noscript","iframe","object","embed","form","button","nav","aside","footer","template","svg":return true
 }
 if attr(n,"hidden")!="" || hasAttr(n,"hidden") || strings.EqualFold(attr(n,"aria-hidden"),"true"){return true}
 if role:=strings.ToLower(attr(n,"role"));role=="dialog"||role=="alertdialog"||role=="navigation"||role=="complementary"||role=="banner"||role=="contentinfo"{return true}
 // Match class/id tokens exactly, not substrings of editorial prose.
 for _,raw:=range []string{attr(n,"class"),attr(n,"id")}{
  for _,word:=range strings.Fields(strings.ToLower(raw)){
   switch word {
   case "ad","ads","advertisement","popup","overlay","newsletter","modal","cookie-banner","paywall-overlay":return true
   }
  }
 }
 return false
}
func hasAttr(n *html.Node,key string)bool{
 if n==nil{return false}
 for _,a:=range n.Attr {if a.Key==key{return true}}
 return false
}
func find(root *html.Node,ok func(*html.Node)bool)[]*html.Node{
 var out []*html.Node
 var walk func(*html.Node)
 walk=func(n *html.Node){
  if n==nil||excluded(n){return}
  if ok(n){out=append(out,n)}
  for c:=n.FirstChild;c!=nil;c=c.NextSibling{walk(c)}
 }
 walk(root)
 return out
}
func plain(n *html.Node)string{
 var b strings.Builder
 var walk func(*html.Node)
 walk=func(x *html.Node){
  if x==nil||excluded(x){return}
  if x.Type==html.TextNode {b.WriteString(x.Data);b.WriteByte(' ');return}
  for c:=x.FirstChild;c!=nil;c=c.NextSibling {walk(c)}
 }
 walk(n)
 return sourcehtml.Normalize(b.String())
}
func safeLinks(n *html.Node,base *url.URL)[]string{
 seen:=map[string]bool{}
 var out []string
 for _,a:=range find(n,func(x *html.Node)bool{return element(x,"a")}){
  resolved:=sourcehtml.ResolveURL(attr(a,"href"),base)
  if resolved!=""&&!seen[resolved] {out=append(out,resolved);seen[resolved]=true}
 }
 return out
}
func creditedCaption(n *html.Node)(string,string){
 // Keep only explicit credit-bearing child metadata. Do not split text on
 // photographer-like words or infer attribution from proximity.
 var credit string
 credits:=find(n,func(x *html.Node)bool{return element(x,"small")||element(x,"cite")})
 if len(credits)==1 {credit=plain(credits[0])}
 cap:=plain(n)
 // No inferred splitting: when a credit is present, caption remains verbatim
 // and attribution is also retained as separately identified evidence.
 return cap,credit
}

// compactImageTextWrapper is *diagnostic* evidence only. It searches at most
// three ancestors below the article for a subtree with exactly one image and
// a separate sibling branch containing text. It does not infer that the text
// is a caption, infer CSS visibility, or authorize removal.
func compactImageTextWrapper(img *html.Node) bool {
 for scope,depth:=img.Parent,0;scope!=nil && depth<3;scope,depth=scope.Parent,depth+1 {
  if element(scope,"article") || element(scope,"main") || element(scope,"body") {break}
  if excluded(scope) {break}
  if len(find(scope,func(x *html.Node)bool{return element(x,"img")}))!=1 {continue}
  for child:=scope.FirstChild;child!=nil;child=child.NextSibling{
   if child.Type!=html.ElementNode || excluded(child) {continue}
   if len(find(child,func(x *html.Node)bool{return element(x,"img")}))!=0 {continue}
   if plain(child)!="" {return true}
  }
 }
 return false
}


// boundedMultiImageGroup is diagnostic-only: within four ancestor edges,
// find a nonarticle container with 2–8 images in at least two distinct
// child branches. It is evidence of a common image group, *not* evidence
// of caption ownership, visibility, or duplicate image identity.
func boundedMultiImageGroup(img *html.Node) bool {
 for scope,depth:=img.Parent,0;scope!=nil && depth<4;scope,depth=scope.Parent,depth+1 {
  if element(scope,"article") || element(scope,"main") || element(scope,"body") {break}
  if excluded(scope) {break}
  images:=find(scope,func(x *html.Node)bool{return element(x,"img")})
  if len(images)<2 || len(images)>8 {continue}
  branches:=0
  for child:=scope.FirstChild;child!=nil;child=child.NextSibling {
   if child.Type!=html.ElementNode || excluded(child) {continue}
   if len(find(child,func(x *html.Node)bool{return element(x,"img")}))>0 {branches++}
  }
  if branches>=2 {return true}
 }
 return false
}

func Inspect(raw,baseURL string)(Report,error){
 report:=Report{Status:"inconclusive"}
 base,err:=url.Parse(baseURL)
 if err!=nil || (base.Scheme!="http"&&base.Scheme!="https") || base.Hostname()=="" || base.User!=nil{
  return report,fmt.Errorf("invalid article URL")
 }
 dom,err:=html.Parse(strings.NewReader(raw))
 if err!=nil{return report,err}
 articles:=find(dom,func(n *html.Node)bool{return element(n,"article")})
 if len(articles)!=1 {return report,nil}
 report.Status="candidate"
 article:=articles[0]
 // Track media wrappers so nested paragraphs are never copied as
 // independent editorial prose.
 var walk func(*html.Node)
 walk=func(n *html.Node){
  if n==nil{return}
  if excluded(n){report.SkippedNoneditorial++;return}
  if n.Type!=html.ElementNode {return}
  if element(n,"figure"){
   images:=find(n,func(x *html.Node)bool{return element(x,"img")})
   caps:=find(n,func(x *html.Node)bool{return element(x,"figcaption")})
   if len(images)!=1 || len(caps)>1 {
    report.AmbiguousFigures++
    if len(images)>1 {report.MultiImageFigures++} else if len(images)==0 {report.InvalidFigureImages++}
    if len(caps)>1 {report.MultipleCaptions++}
    return
   }
   imageURL:=sourcehtml.ResolveURL(attr(images[0],"src"),base)
   if imageURL=="" {report.AmbiguousFigures++;report.InvalidFigureImages++;return}
   item:=Item{Kind:"figure",ImageURL:imageURL}
   if len(caps)==1 {item.Caption,item.Credit=creditedCaption(caps[0])}
   report.Items=append(report.Items,item)
   return
  }
  switch n.Data {
  case "p","h1","h2","h3","h4","h5","h6":
   // Only leaf-level block paragraphs/headings are observed. A paragraph
   // nested in a figure has already been consumed by the figure branch.
   if s:=plain(n);s!=""{report.Items=append(report.Items,Item{Kind:n.Data,Text:s,Links:safeLinks(n,base)})}
   return
  case "img":
   // Only images with explicit figure association are currently credited.
   // Detached images are marked ambiguous rather than guessed into prose.
   report.AmbiguousFigures++
   report.UnwrappedImages++
   if compactImageTextWrapper(n) {report.CompactImageTextWrappers++} else {report.OtherUnwrappedImages++; if boundedMultiImageGroup(n) {report.OtherInMultiImageGroups++} else {report.OtherWithoutBoundedGroup++}}
   return
  }
  for c:=n.FirstChild;c!=nil;c=c.NextSibling{walk(c)}
 }
 walk(article)
 return report,nil
}
