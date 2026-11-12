package sourcepreserve

import (
 "fmt"
 "net/url"
 "strings"

 "golang.org/x/net/html"
)

// ImageTrace is an offline structural diagnostic. All identifiers are
// per-execution ordinals. No publisher URLs, text, classes or IDs are emitted.
type ImageTrace struct {
 ImageIndex int `json:"imageIndex"`
 Pattern string `json:"pattern"`
 ParentTag string `json:"parentTag"`
 GroupIndex int `json:"groupIndex,omitempty"`
 GroupImages int `json:"groupImages,omitempty"`
 GroupImageBranches int `json:"groupImageBranches,omitempty"`
 GroupTextBranches int `json:"groupTextBranches,omitempty"`
 Ancestors []AncestorTrace `json:"ancestors,omitempty"`
}
// AncestorTrace records a bounded ancestor's *direct child* branches. A
// text-bearing branch is not proof of caption ownership.
type AncestorTrace struct {
 Depth int `json:"depth"`
 Tag string `json:"tag"`
 Images int `json:"images"`
 ImageBranches int `json:"imageBranches"`
 TextBranchesBefore int `json:"textBranchesBefore,omitempty"`
 TextBranchesAfter int `json:"textBranchesAfter,omitempty"`
 ExplicitCaptionElements int `json:"explicitCaptionElements,omitempty"`
}
type StructureTrace struct {
 Status string `json:"status"`
 Images []ImageTrace `json:"images,omitempty"`
}

// safeTraceTag prevents arbitrary custom-element names from becoming a
// content channel. Only a fixed vocabulary is shown.
func safeTraceTag(s string)string{
 switch s {
 case "a","div","span","section","picture","figure","figcaption","p","small","cite","article","main","body","header","footer","ul","li","ol","table","td","th","tr","aside":
  return s
 }
 return "other"
}

// traceAncestors records up to six enclosing elements, stopping before the
// article root. Counts reflect direct child branches, not rendered layout.
func traceAncestors(img *html.Node) []AncestorTrace {
 var result []AncestorTrace
 for scope,depth:=img.Parent,1;scope!=nil && depth<=6;scope,depth=scope.Parent,depth+1 {
  if element(scope,"article")||element(scope,"main")||element(scope,"body"){break}
  if excluded(scope){break}
  item:=AncestorTrace{Depth:depth,Tag:safeTraceTag(scope.Data)}
  item.Images=len(find(scope,func(x *html.Node)bool{return element(x,"img")}))
  item.ExplicitCaptionElements=len(find(scope,func(x *html.Node)bool{return element(x,"figcaption")}))
  // Determine which immediate child branch contains this image.
  var imageBranch *html.Node
  for branch:=img;branch!=nil&&branch!=scope;branch=branch.Parent{
   if branch.Parent==scope {imageBranch=branch;break}
  }
  after:=false
  for child:=scope.FirstChild;child!=nil;child=child.NextSibling {
   if child==imageBranch {after=true}
   if child.Type!=html.ElementNode||excluded(child){continue}
   if len(find(child,func(x *html.Node)bool{return element(x,"img")}))>0 {
    item.ImageBranches++
   } else if plain(child)!="" {
    if after {item.TextBranchesAfter++} else {item.TextBranchesBefore++}
   }
  }
  result=append(result,item)
 }
 return result
}

// TraceStructure observes exactly the unwrapped images visited by Inspect.
// It does not make network requests, export content, infer caption ownership,
// or alter the source HTML or extracted Items.
func TraceStructure(raw,baseURL string)(StructureTrace,error){
 result:=StructureTrace{Status:"inconclusive"}
 u,err:=url.Parse(baseURL)
 if err!=nil || (u.Scheme!="https"&&u.Scheme!="http") || u.Hostname()=="" || u.User!=nil {
  return result,fmt.Errorf("invalid article URL")
 }
 dom,err:=html.Parse(strings.NewReader(raw))
 if err!=nil {return result,err}
 articles:=find(dom,func(n *html.Node)bool{return element(n,"article")})
 if len(articles)!=1 {return result,nil}
 result.Status="candidate"
 groups:=map[*html.Node]int{}
 var walk func(*html.Node)
 walk=func(n *html.Node){
  if n==nil||excluded(n)||n.Type!=html.ElementNode {return}
  if element(n,"figure"){return}
  switch n.Data {
  case "p","h1","h2","h3","h4","h5","h6":return
  case "img":
   item:=ImageTrace{ImageIndex:len(result.Images)+1,Pattern:"unwrapped",ParentTag:"unknown"}
   if n.Parent!=nil {item.ParentTag=safeTraceTag(n.Parent.Data)}
   item.Ancestors=traceAncestors(n)
   if compactImageTextWrapper(n) {item.Pattern="compact-text-sibling"} else if group:=boundedMultiImageGroup(n);group!=nil {
    item.Pattern="bounded-multi-image-group"
    if groups[group]==0 {groups[group]=len(groups)+1}
    item.GroupIndex=groups[group]
    // Group metrics are deliberately only structural. They are not a
    // guarantee that the group is a single publisher gallery.
    item.GroupImages=len(find(group,func(x *html.Node)bool{return element(x,"img")}))
    for child:=group.FirstChild;child!=nil;child=child.NextSibling{
     if child.Type!=html.ElementNode || excluded(child) {continue}
     if len(find(child,func(x *html.Node)bool{return element(x,"img")}))>0 {
      item.GroupImageBranches++
     } else if plain(child)!="" {item.GroupTextBranches++}
    }
   }
   result.Images=append(result.Images,item)
   return
  }
  for child:=n.FirstChild;child!=nil;child=child.NextSibling{walk(child)}
 }
 walk(articles[0])
 return result,nil
}
