package main

import (
 "fmt"

 "clean-article-reader/internal/sourcehtml"
)

// photoScope is a diagnostic relationship, not a license to delete a block.
// A "single image" scope may still contain real editorial text.
type photoScope struct {
 MatchPath string
 PhotoOrdinal int
 ScopeTag string
 ScopePath string
 Edges int
 GroupImages int
}

func collectPhotoScopes(raw, passage string) []photoScope {
 root:=sourcehtml.Parse(raw)
 byPath:=map[string]*sourcehtml.Node{}
 paths:=map[*sourcehtml.Node]string{}
 imageOrdinal:=map[*sourcehtml.Node]int{}
 nextImage:=0
 var index func(*sourcehtml.Node,string)
 index=func(n *sourcehtml.Node,path string) {
  byPath[path]=n
  paths[n]=path
  if n.Tag=="img" {nextImage++;imageOrdinal[n]=nextImage}
  counters:=map[string]int{}
  for _,child:=range n.Kids {
   if child.Tag=="#text" {continue}
   counters[child.Tag]++
   index(child,fmt.Sprintf("%s/%s[%d]",path,child.Tag,counters[child.Tag]))
  }
 }
 index(root,"root")
 // Cache image descendants; no URL, class, image alt text or source content
 // is ever printed or stored in a report.
 imageNodes:=map[*sourcehtml.Node][]*sourcehtml.Node{}
 var images func(*sourcehtml.Node) []*sourcehtml.Node
 images=func(n *sourcehtml.Node) []*sourcehtml.Node {
  if found,ok:=imageNodes[n];ok{return found}
  found:=[]*sourcehtml.Node{}
  if n.Tag=="img" {found=append(found,n)}
  for _,child:=range n.Kids {
   if child.Tag=="#text" || sourcehtml.BlockedTags[child.Tag] {continue}
   found=append(found,images(child)...)
  }
  imageNodes[n]=found
  return found
 }
 matches:=scanProvenance(raw,passage)
 result:=make([]photoScope,0,len(matches))
 for _,match:=range matches {
  n:=byPath[match.Path]
  if n==nil {continue}
  resultItem:=photoScope{MatchPath:match.Path,ScopeTag:"none",ScopePath:"none"}
  edges:=0
  for anc:=n;anc!=nil;anc=anc.Parent {
   if anc.Tag=="article" || anc.Tag=="main" || anc.Tag=="body" || anc.Tag=="root" {break}
   containing:=images(anc)
   if len(containing)==1 {
    resultItem.PhotoOrdinal=imageOrdinal[containing[0]]
    resultItem.ScopeTag=anc.Tag
    resultItem.ScopePath=paths[anc]
    resultItem.Edges=edges
    resultItem.GroupImages=1
    break
   }
   edges++
  }
  // If no distinct single-image subtree exists, report the nearest group
  // without suggesting which image owns a paragraph.
  if resultItem.PhotoOrdinal==0 {
   for anc:=n;anc!=nil;anc=anc.Parent {
    if anc.Tag=="article" || anc.Tag=="main" || anc.Tag=="body" || anc.Tag=="root" {break}
    count:=len(images(anc))
    if count>=2 && count<=12 {
     resultItem.GroupImages=count
     resultItem.ScopeTag=anc.Tag
     resultItem.ScopePath=paths[anc]
     break
    }
   }
  }
  result=append(result,resultItem)
 }
 return result
}

func renderPhotoScopes(raw,passage string) {
 items:=collectPhotoScopes(raw,passage)
 fmt.Printf("    publisher photo scopes: %d matches\n",len(items))
 const limit=8
 for i,s:=range items {
  if i>=limit {fmt.Printf("      ... %d further scopes omitted\n",len(items)-limit);break}
  if s.PhotoOrdinal>0 {
   fmt.Printf("      match %d image#%d scope=%s ancestor-edges=%d scope-path=%s\n",
    i+1,s.PhotoOrdinal,s.ScopeTag,s.Edges,s.ScopePath)
  } else if s.GroupImages>0 {
   fmt.Printf("      match %d multi-image-group=%d scope=%s scope-path=%s (image owner unknown)\n",
    i+1,s.GroupImages,s.ScopeTag,s.ScopePath)
  } else {
   fmt.Printf("      match %d no image ancestor before article boundary\n",i+1)
  }
 }
 fmt.Println("    Photo ordinals are local to the source DOM, not image asset identities.")
}

