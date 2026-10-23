package recovery

import (
 "fmt"
 "strings"

 "clean-article-reader/internal/media"
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/sourcehtml"
)

// sourceGalleryCandidate identifies a compact image-centric layout. It uses
// source structure and CSS layout semantics, never publisher-specific classes.
func SourceGalleryCandidate(n *sourcehtml.Node) bool {
 if n.Tag!="div" && n.Tag!="section" && n.Tag!="figure" && n.Tag!="ul" {return false}
 identity:=strings.ToLower(n.Attr["class"]+" "+n.Attr["id"]+" "+n.Attr["data-gallery"])
 explicit:=strings.Contains(identity,"gallery")||strings.Contains(identity,"carousel")||strings.Contains(identity,"slideshow")
 style:=strings.ToLower(n.Attr["style"])
 isFlex:=strings.Contains(style,"display:flex")||strings.Contains(style,"display: flex")
 isGrid:=strings.Contains(style,"display:grid")||strings.Contains(style,"display: grid")
 verticalFlex:=strings.Contains(style,"flex-direction:column")||strings.Contains(style,"flex-direction: column")
 horizontal:=(isFlex&&!verticalFlex)||isGrid
 for _,token:=range strings.Fields(identity) {
  if token=="flex-row"||strings.HasSuffix(token,":flex-row")||strings.HasPrefix(token,"grid-cols-")||strings.Contains(token,":grid-cols-") {horizontal=true}
 }
 if !explicit&&!horizontal {return false}
 // Require multiple distinct image-centric immediate branches, not an article
 // wrapper with interleaved prose, buttons, or unrelated body sections.
 count:=0
 for _,child:=range n.Kids {
  if child.Tag=="#text" {continue}
  if child.Tag=="p"||child.Tag=="h1"||child.Tag=="h2"||child.Tag=="h3" {return false}
  images:=sourcehtml.Find(child,func(x *sourcehtml.Node)bool{return x.Tag=="img"})
  if len(images)==0 {return false}
  // Prevent a whole grid of article sections from looking like a photo group.
  if len(images)!=1 {return false}
  if len(sourcehtml.Find(child,func(x *sourcehtml.Node)bool{return x.Tag=="p"||x.Tag=="article"}))>0 {return false}
  count++
 }
 return count>=2 && count<=12
}

// groupArticleGalleryImages restores photo layouts using publisher-neutral
// structural evidence. Editorial images without a compact shared layout stay
// independent, regardless of their proximity in the extracted text.
func GroupArticleGalleryImages(a *article.Article, sourceHTML, baseURL string, resolveImage func(*sourcehtml.Node,string) string) {
 if a==nil||len(a.Blocks)==0{return}
 root:=sourcehtml.Parse(sourceHTML)
 groups:=[][]string{}
 claimed:=map[string]bool{}
 // Walk deepest-first so nested layout wrappers win over broader containers.
 candidates:=sourcehtml.Find(root,SourceGalleryCandidate)
 for i:=len(candidates)-1;i>=0;i-- {
  n:=candidates[i]
  urls:=[]string{}
  seen:=map[string]bool{}
  for _,img:=range sourcehtml.Find(n,func(x *sourcehtml.Node)bool{return x.Tag=="img"}) {
   u:=resolveImage(img,baseURL)
   key:=media.Identity(u)
   if u==""||key==""||seen[key]||claimed[key] {continue}
   seen[key]=true
   urls=append(urls,u)
  }
  if len(urls)<2||len(urls)>12 {continue}
  groups=append(groups,urls)
  for _,u:=range urls {claimed[media.Identity(u)]=true}
 }
 // Match by canonical media asset identity so source and Readability
 // renditions can differ in image dimensions.
 byIdentity:=map[string]int{}
 for i,b:=range a.Blocks {if b.Kind=="img"&&b.Role!="lead"&&b.Role!="author" {byIdentity[media.Identity(b.URL)]=i}}
 idFor:=map[int]string{}
 for gi,urls:=range groups {
  matches:=[]int{}
  seen:=map[int]bool{}
  for _,u:=range urls {if i,ok:=byIdentity[media.Identity(u)];ok&&!seen[i] {seen[i]=true;matches=append(matches,i)}}
  if len(matches)<2 {continue}
  id:=fmt.Sprintf("gallery-%d",gi+1)
  for _,i:=range matches {idFor[i]=id}
 }
 if len(idFor)==0{return}
 members:=map[string][]article.Block{}
 first:=map[string]int{}
 for i,b:=range a.Blocks {if id,ok:=idFor[i];ok {
  b.Gallery=id
  members[id]=append(members[id],b)
  if _,exists:=first[id];!exists {first[id]=i}
 }}
 out:=make([]article.Block,0,len(a.Blocks))
 for i,b:=range a.Blocks {
  if id,ok:=idFor[i];ok {
   if first[id]==i {out=append(out,members[id]...)}
   continue
  }
  out=append(out,b)
 }
 a.Blocks=out
}
