package sourcepreserve

import (
 "fmt"
 "net/url"
 "strings"

 "golang.org/x/net/html"
)

// ImageOccurrence describes one observed image in DOM traversal order.
// SourceURL is used only in memory; it is never included in the report.
type ImageOccurrence struct {
 SourceURL string `json:"-"`
 Ordinal int `json:"-"`
 ExplicitFigure bool `json:"-"`
 ExplicitCaption bool `json:"-"`
}

// ImageMatch is a one-to-one exact resolved URL match, not proof of visual
// identity or equivalence of responsive asset variants.
type ImageMatch struct {
 SourceIndex int `json:"sourceIndex"`
 SelectedIndex int `json:"selectedIndex"`
}

type ImageAlignment struct {
 SourceStatus string `json:"sourceStatus"`
 SourceImages int `json:"sourceImages"`
 SelectedImages int `json:"selectedImages"`
 ExactURLMatches int `json:"exactURLMatches"`
 SourceOnly int `json:"sourceOnly"`
 SelectedOnly int `json:"selectedOnly"`
 SourceExplicitFigureImages int `json:"sourceExplicitFigureImages,omitempty"`
 SelectedExplicitFigureImages int `json:"selectedExplicitFigureImages,omitempty"`
 SourceExplicitCaptionImages int `json:"sourceExplicitCaptionImages,omitempty"`
 SelectedExplicitCaptionImages int `json:"selectedExplicitCaptionImages,omitempty"`
 Matches []ImageMatch `json:"matches,omitempty"`
 OrderPreserved bool `json:"orderPreserved"`
}

// selectedImages walks a semantic tree to preserve image order and records
// only an explicit figure containing exactly one image and one figcaption.
// Sibling prose or wrapper proximity does NOT imply image attribution.
func selectedImages(nodes []SemanticNode) []ImageOccurrence {
 var out []ImageOccurrence
 var walk func([]SemanticNode)
 walk=func(items []SemanticNode){
  for _,n:=range items{
   if n.Kind=="image" {
    out=append(out,ImageOccurrence{SourceURL:n.URL,Ordinal:len(out)+1})
    continue
   }
   if n.Kind=="figure" {
    before:=len(out)
    walk(n.Children)
    if len(out)==before+1 {
     out[before].ExplicitFigure=true
     captions:=0
     for _,c:=range n.Children {if c.Kind=="figcaption"{captions++}}
     out[before].ExplicitCaption=captions==1
    }
   } else {walk(n.Children)}
  }
 }
 walk(nodes)
 return out
}

// sourceArticleImages requires exactly one explicit source article scope.
// A page with zero/multiple article nodes is inconclusive, not a valid count
// of zero images. This is source corroboration, NOT source article selection.
func sourceArticleImages(raw,baseURL string)([]ImageOccurrence,string,error){
 u,err:=url.Parse(baseURL)
 if err!=nil || (u.Scheme!="http"&&u.Scheme!="https") || u.Hostname()=="" || u.User!=nil{return nil,"inconclusive",fmt.Errorf("invalid HTTP(S) base URL")}
 dom,err:=html.Parse(strings.NewReader(raw));if err!=nil{return nil,"inconclusive",err}
 scopes:=find(dom,func(n *html.Node)bool{return element(n,"article")})
 if len(scopes)!=1{return nil,"inconclusive",nil}
 // Convert only the explicit article subtree through the same safe semantic
 // rules as selected HTML; do not search adjacent sidebars or whole page.
 doc:=SemanticDocument{}
 nodes:=convertNode(scopes[0],u,&doc)
 return selectedImages(nodes),"unique-article",nil
}

// CompareImageAlignment matches source-article occurrences against
// Readability-selected semantic images on exact resolved HTTP(S) src only.
// Repeated URLs are paired in occurrence order. The JSON contains no URLs,
// publisher prose, HTML or attributes. Missing source scope is inconclusive.
func CompareImageAlignment(sourceHTML,baseURL string,selected SemanticDocument)(ImageAlignment,error){
 result:=ImageAlignment{SourceStatus:"inconclusive"}
 source,status,err:=sourceArticleImages(sourceHTML,baseURL)
 if err!=nil{return result,err}
 result.SourceStatus=status
 targets:=selectedImages(selected.Nodes)
 result.SelectedImages=len(targets)
 for _,n:=range targets {
  if n.ExplicitFigure {result.SelectedExplicitFigureImages++}
  if n.ExplicitCaption {result.SelectedExplicitCaptionImages++}
 }
 if status!="unique-article"{return result,nil}
 result.SourceImages=len(source)
 positions:=map[string][]int{}
 for _,n:=range source{
  positions[n.SourceURL]=append(positions[n.SourceURL],n.Ordinal)
  if n.ExplicitFigure {result.SourceExplicitFigureImages++}
  if n.ExplicitCaption {result.SourceExplicitCaptionImages++}
 }
 last:=0
 result.OrderPreserved=true
 for _,n:=range targets{
  queue:=positions[n.SourceURL]
  if len(queue)==0 {result.SelectedOnly++;continue}
  index:=queue[0]
  positions[n.SourceURL]=queue[1:]
  result.Matches=append(result.Matches,ImageMatch{SourceIndex:index,SelectedIndex:n.Ordinal})
  result.ExactURLMatches++
  if index<last {result.OrderPreserved=false}
  last=index
 }
 result.SourceOnly=len(source)-result.ExactURLMatches
 return result,nil
}
