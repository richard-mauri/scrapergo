package main

import (
 "fmt"
 "clean-article-reader/internal/media"
 "strings"
)

// groupArticleGalleryImages preserves explicitly grouped source galleries.
// Only a container with gallery-related markup and at least two image members
// qualifies. Independent editorial photographs are never grouped by proximity.
func groupArticleGalleryImages(a *Article, sourceHTML, baseURL string) {
 if a == nil || len(a.Blocks)==0 { return }
 root:=parseHTML(sourceHTML)
 groups:=[][]string{}
 claimed:=map[string]bool{}
 for _,n:=range find(root,func(n *Node)bool {
  if n.Tag!="div" && n.Tag!="section" && n.Tag!="figure" && n.Tag!="ul" {return false}
  identity:=strings.ToLower(n.Attr["class"]+" "+n.Attr["id"]+" "+n.Attr["data-gallery"])
  return strings.Contains(identity,"gallery") || strings.Contains(identity,"carousel") || strings.Contains(identity,"slideshow")
 }) {
  // Prefer the innermost explicit group; avoid grouping nested carousels twice.
  nested:=false
  for _,desc:=range find(n,func(x *Node)bool {
   if x==n || (x.Tag!="div"&&x.Tag!="section"&&x.Tag!="figure"&&x.Tag!="ul"){return false}
   cls:=strings.ToLower(x.Attr["class"]+" "+x.Attr["id"])
   return strings.Contains(cls,"gallery")||strings.Contains(cls,"carousel")||strings.Contains(cls,"slideshow")
  }) {if desc!=nil {nested=true;break}}
  if nested {continue}
  urls:=[]string{}
  unique:=map[string]bool{}
  for _,img:=range find(n,func(x *Node)bool{return x.Tag=="img"}) {
   u:=articleImageURL(img,baseURL)
   if u!="" && !unique[u]&&!claimed[media.Identity(u)] {unique[u]=true;urls=append(urls,u)}
  }
  if len(urls)<2||len(urls)>12 {continue}
  groups=append(groups,urls)
  for _,u:=range urls {claimed[media.Identity(u)]=true}
 }
 // PhotoSwipe stores image slides as pswp__item elements, often alongside
 // a separate lightbox rendition. Group only siblings within the same
 // PhotoSwipe track and deduplicate by source-asset identity.
 for _,track:=range find(root,func(n *Node)bool {
  if n.Tag!="div" {return false}
  cls:=" "+strings.ToLower(n.Attr["class"])+" "
  return strings.Contains(cls," pswp__container ") || strings.Contains(cls," pswp__items ")
 }) {
  urls:=[]string{}
  unique:=map[string]bool{}
  for _,child:=range track.Kids {
   if !strings.Contains(" "+strings.ToLower(child.Attr["class"])+" "," pswp__item ") {continue}
   for _,img:=range find(child,func(n *Node)bool{return n.Tag=="img"}) {
    u:=articleImageURL(img,baseURL)
    key:=media.Identity(u)
    if u!=""&&key!=""&&!unique[key]&&!claimed[key] {
     unique[key]=true
     urls=append(urls,u)
     break
    }
   }
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
 members:=map[string][]Block{}
 first:=map[string]int{}
 for i,b:=range a.Blocks {if id,ok:=idFor[i];ok {
  b.Gallery=id
  members[id]=append(members[id],b)
  if _,exists:=first[id];!exists {first[id]=i}
 }}
 out:=make([]Block,0,len(a.Blocks))
 for i,b:=range a.Blocks {
  if id,ok:=idFor[i];ok {
   if first[id]==i {out=append(out,members[id]...)}
   continue
  }
  out=append(out,b)
 }
 a.Blocks=out
}
