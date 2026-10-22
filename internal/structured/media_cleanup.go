// Package structured handles HTML and block-level structural conversion.
package structured

import (
 "bytes"
 "regexp"
 "strings"

 "clean-article-reader/internal/article"
 "golang.org/x/net/html"
)

var escapedMediaElementRE = regexp.MustCompile(`(?is)<\s*/?\s*(?:iframe|audio|video|embed|object)\b[^>]*>`)

// stripEmbeddedMedia removes actual non-editorial media DOM nodes before
// Readability serializes HTML to text. Sibling article prose is preserved.
func StripEmbeddedMedia(raw string) string {
 root,err:=html.Parse(strings.NewReader(raw))
 if err!=nil{return raw}
 var visit func(*html.Node)
 visit=func(n *html.Node){
  if n.Type==html.TextNode && escapedMediaElementRE.MatchString(n.Data) {
    insideCode:=false
    for p:=n.Parent;p!=nil;p=p.Parent { if p.Type==html.ElementNode&&(p.Data=="pre"||p.Data=="code"){insideCode=true;break} }
    if !insideCode {n.Data=escapedMediaElementRE.ReplaceAllString(n.Data,"")}
  }
  for child:=n.FirstChild;child!=nil;{
   next:=child.NextSibling
   if child.Type==html.ElementNode && (child.Data=="iframe"||child.Data=="embed"||child.Data=="audio"||child.Data=="video"||child.Data=="object"){
    n.RemoveChild(child)
   }else{visit(child)}
   child=next
  }
 }
 visit(root)
 var b bytes.Buffer
 if err:=html.Render(&b,root);err!=nil{return raw}
 return b.String()
}

// removeSerializedPlayerBlocks rejects standalone serialized player widgets
// that survived Readability as article blocks. It does not inspect neighboring
// blocks, and never removes a list item just because it says "Embed".
func RemoveSerializedPlayerBlocks(a *article.Article) {
 if a==nil {return}
 out:=make([]article.Block,0,len(a.Blocks))
 for _,b:=range a.Blocks {
  if b.Kind!="img" && containsSerializedMediaTag(b.Text) {
   // Only suppress blocks that are player widgets rather than mixed prose.
   remainder:=escapedMediaElementRE.ReplaceAllString(b.Text,"")
   remainder=strings.TrimSpace(strings.ToLower(remainder))
   remainder=strings.TrimSpace(strings.TrimPrefix(remainder,"embed"))
   remainder=strings.TrimSpace(remainder)
   if remainder=="" {continue}
  }
  out=append(out,b)
 }
 a.Blocks=out
}
func containsSerializedMediaTag(text string) bool {
 return escapedMediaElementRE.MatchString(text)
}

