package main

import (
 "bytes"
 "fmt"
 "html"
 "net/url"
 "sort"
 "strings"
 "unicode"

 "clean-article-reader/internal/article"
 "clean-article-reader/internal/sourcehtml"

 readability "codeberg.org/readeck/go-readability/v2"
)

// provenanceNode describes where a passage was found, without recording or
// printing publisher text, URLs, class names, IDs, or attribute values.
type provenanceNode struct {
 Order int
 Path string
 Parent string
 Container string
 Tag string
 Origin string
 Mode string
}

// comparableHTMLText compares publisher text across inline markup boundaries.
// "compact" matches are weaker evidence: whitespace/entity serialization can
// differ, and multiple origins must never be inferred from a compact match.
func comparableHTMLText(s string) string {
 return normalize(html.UnescapeString(s))
}
func compactHTMLText(s string) string {
 return strings.Map(func(r rune) rune {
  if unicode.IsSpace(r) {return -1}
  return r
 },comparableHTMLText(s))
}
func containsCandidate(haystack, target string) string {
 h,t:=comparableHTMLText(haystack),comparableHTMLText(target)
 if h=="" || t=="" {return ""}
 if strings.Contains(h,t) {return "text"}
 a,b:=compactHTMLText(haystack),compactHTMLText(target)
 if len([]rune(b))>=24 && strings.Contains(a,b) {return "compact"}
 return ""
}

// scanProvenance records only minimal matching elements. A container match is
// not reported when one of its descendants already matches the same passage.
// All paths and ordinal positions are computed independently for each DOM tree.
func scanProvenance(raw, passage string) []provenanceNode {
 root:=sourcehtml.Parse(raw)
 out:=[]provenanceNode{}
 order:=0
 var walk func(*sourcehtml.Node,string,string,string) (string,bool)
 walk=func(n *sourcehtml.Node,path,parent,container string)(string,bool) {
  if n==nil || sourcehtml.BlockedTags[n.Tag] {return "",false}
  if n.Tag=="#text" {return n.Text,false}
  if n.Tag=="root" {path="root"}
  if n.Tag!="root" {order++}
  myOrder:=order
  nearest:=container
  switch n.Tag {
  case "article","main","section","figure","figcaption","picture","blockquote","li":
   nearest=n.Tag
  }
  var all strings.Builder
  matchedChild:=false
  counters:=map[string]int{}
  for _,child:=range n.Kids {
   if child.Tag=="#text" {
    all.WriteString(child.Text)
    continue
   }
   counters[child.Tag]++
   childPath:=fmt.Sprintf("%s/%s[%d]",path,child.Tag,counters[child.Tag])
   content,matched:=walk(child,childPath,n.Tag,nearest)
   all.WriteString(content)
   if matched {matchedChild=true}
  }
  mode:=containsCandidate(all.String(),passage)
  if mode!="" && !matchedChild && n.Tag!="root" {
   out=append(out,provenanceNode{
    Order:myOrder,Path:path,Parent:parent,Container:nearest,
    Tag:n.Tag,Origin:"text",Mode:mode,
   })
  }
  // Attributes represent distinct potential origins, most importantly
  // photograph alt/title text. Only record the attribute NAME, never VALUE.
  for _,attr:=range []string{"alt","title","aria-label","data-caption"} {
   if value:=n.Attr[attr];value!="" {
    if mode:=containsCandidate(value,passage);mode!="" {
     out=append(out,provenanceNode{
      Order:myOrder,Path:path,Parent:parent,Container:nearest,
      Tag:n.Tag,Origin:"attr:"+attr,Mode:mode,
     })
    }
   }
  }
  return all.String(),mode!="" || matchedChild
 }
 walk(root,"root","","")
 // A parent may be examined before a matching child in a separate branch;
 // sort by source document order for stable evidence.
 sort.SliceStable(out,func(i,j int)bool {
  if out[i].Order!=out[j].Order {return out[i].Order<out[j].Order}
  if out[i].Path!=out[j].Path {return out[i].Path<out[j].Path}
  return out[i].Origin<out[j].Origin
 })
 return out
}

// renderProvenance is deliberately bounded; no source or article content can
// appear in its output. Truncation is explicit rather than silently misleading.
func renderProvenance(label string,nodes []provenanceNode) {
 fmt.Printf("    %s: %d minimal matches\n",label,len(nodes))
 const max=8
 for i,n:=range nodes {
  if i>=max {fmt.Printf("      ... %d further matches omitted\n",len(nodes)-max);break}
  fmt.Printf("      #%d %s parent=%s context=%s origin=%s mode=%s\n",
   n.Order,n.Path,n.Parent,n.Container,n.Origin,n.Mode)
 }
}

// reportSourceProvenance evaluates both blocks separately, not only their
// shorter shared text; this helps distinguish a caption from its credit.
func reportSourceProvenance(raw,sourceURL string,blocks []article.Block,pairs []match) error {
 base,err:=url.Parse(sourceURL)
 if err!=nil || !base.IsAbs() || (base.Scheme!="http" && base.Scheme!="https") {
  return fmt.Errorf("base URL must be absolute HTTP(S)")
 }
 parsed,err:=readability.FromReader(strings.NewReader(raw),base)
 if err!=nil {return fmt.Errorf("Readability parse: %w",err)}
 var selected bytes.Buffer
 if err:=parsed.RenderHTML(&selected);err!=nil {return fmt.Errorf("Readability RenderHTML: %w",err)}
 fmt.Println("Element provenance (1-based final blocks; source and selected HTML document-order paths; no content)")
 for _,p:=range pairs {
  if p.First<1 || p.Second<1 || p.First>len(blocks) || p.Second>len(blocks) {continue}
  fmt.Printf("  pair %d/%d (%s):\n",p.First,p.Second,p.Type)
  // Report both unique texts; for identical pairs print just one. The
  // matching element count can legitimately be two in original source.
  indexes:=[]int{p.First}
  if comparableHTMLText(blocks[p.First-1].Text)!=comparableHTMLText(blocks[p.Second-1].Text) {
   indexes=append(indexes,p.Second)
  }
  for _,index:=range indexes {
   passage:=blocks[index-1].Text
   fmt.Printf("    canonical block %d (%s):\n",index,blocks[index-1].Kind)
   renderProvenance("publisher source",scanProvenance(raw,passage))
   renderProvenance("Readability HTML",scanProvenance(selected.String(),passage))
  }
 }
 fmt.Println("Paths are local to each separate HTML tree. Compact matches ignore whitespace; neither matches nor ordinals prove one-to-one source identity.")
 fmt.Println("Source element ancestry helps locate representations, but Readability does not supply a source-node identity map.")
 return nil
}
