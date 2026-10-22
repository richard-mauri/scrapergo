package main

import (
 "encoding/json"
 "io"
 "net"
 "net/http"
 "net/url"
 "strings"
)

type traceImage struct {
 Stage string `json:"stage"`
 Index int `json:"index"`
 URL string `json:"url"`
 Role string `json:"role,omitempty"`
 Caption string `json:"caption,omitempty"`
 Parent string `json:"parent,omitempty"`
 ParentClass string `json:"parentClass,omitempty"`
 Grandparent string `json:"grandparent,omitempty"`
 GrandparentClass string `json:"grandparentClass,omitempty"`
 SourceIdentity string `json:"sourceIdentity,omitempty"`
}
func traceImageBlocks(stage string, blocks []Block) []traceImage {
 out:=[]traceImage{}
 for i,b:=range blocks {
  if b.Kind!="img" {continue}
  out=append(out,traceImage{Stage:stage,Index:i,URL:b.URL,Role:b.Role,Caption:b.Text,SourceIdentity:imageSourceIdentity(b.URL)})
 }
 return out
}
func sourceImageTrace(raw,articleURL string) []traceImage {
 root:=parseHTML(raw)
 out:=[]traceImage{}
 for _,n:=range find(root,func(n *Node)bool{return n.Tag=="img"}) {
  parent,gparent:="",""
  pc,gc:="",""
  if n.Parent!=nil {parent=n.Parent.Tag;pc=n.Parent.Attr["class"];if n.Parent.Parent!=nil{gparent=n.Parent.Parent.Tag;gc=n.Parent.Parent.Attr["class"]}}
  src:=articleImageURL(n,articleURL)
  if src=="" {continue}
  out=append(out,traceImage{Stage:"source-dom",Index:len(out),URL:src,Caption:n.Attr["alt"],Parent:parent,ParentClass:pc,Grandparent:gparent,GrandparentClass:gc,SourceIdentity:imageSourceIdentity(src)})
  if len(out)>=150 {break}
 }
 return out
}
// diagnosticImageTrace is read-only and restricted to direct loopback peers.
// It traces the same transformations as writeExtractedArticle, without
// returning raw publisher HTML, credentials, or making asset requests.
func diagnosticImageTrace(w http.ResponseWriter,r *http.Request) {
 w.Header().Set("Content-Type","application/json; charset=utf-8")
 w.Header().Set("Cache-Control","no-store")
 if r.Method!=http.MethodPost {w.Header().Set("Allow","POST");http.Error(w,"POST required",http.StatusMethodNotAllowed);return}
 peer,_,err:=net.SplitHostPort(r.RemoteAddr)
 if err!=nil||net.ParseIP(peer)==nil||!net.ParseIP(peer).IsLoopback(){http.Error(w,"localhost only",http.StatusForbidden);return}
 r.Body=http.MaxBytesReader(w,r.Body,maxHTML+8192)
 var input struct {URL string `json:"url"`; HTML string `json:"html"`}
 dec:=json.NewDecoder(r.Body)
 if err:=dec.Decode(&input);err!=nil{http.Error(w,"invalid JSON",400);return}
 var trailing any
 if err:=dec.Decode(&trailing);err!=io.EOF {http.Error(w,"one JSON object required",400);return}
 u,err:=url.Parse(input.URL)
 if err!=nil|| (u.Scheme!="http"&&u.Scheme!="https")||u.Hostname()==""||u.User!=nil||u.Fragment!=""||strings.ContainsAny(input.URL,"\r\n")||len(input.HTML)>maxHTML{http.Error(w,"invalid URL or HTML",400);return}
 if input.HTML=="" {
  resolved,err:=resolveGoogleNewsURL(r.Context(),input.URL)
  if err!=nil {http.Error(w,"URL resolution failed",422);return}
  raw,final,err:=download(r.Context(),resolved)
  if err!=nil {http.Error(w,"article retrieval failed",422);return}
  input.HTML=raw
  input.URL=final
 }
 cleaned:=stripEmbeddedMedia(input.HTML)
 article,err:=extractWithReadability(cleaned,input.URL)
 if err!=nil{http.Error(w,"extraction failed",422);return}
 stages:=[]traceImage{}
 stages=append(stages,sourceImageTrace(cleaned,input.URL)...)
 stages=append(stages,traceImageBlocks("readability",article.Blocks)...)
 recoverArticlePhotographs(&article,cleaned,input.URL)
 deduplicateArticleImages(&article)
 recoverMissingArticleLinks(&article,cleaned,input.URL)
 recoverAuthorPortrait(&article,cleaned,input.URL)
 stages=append(stages,traceImageBlocks("recovered",article.Blocks)...)
 removeSerializedPlayerBlocks(&article)
 classifyContributorImages(&article)
 classifyImagesFromSource(&article,cleaned,input.URL)
 reconcileContributorByline(&article)
 separateContributorHeader(&article)
 stages=append(stages,traceImageBlocks("classified",article.Blocks)...)
 sourceLead:=""
 if base,err:=url.Parse(input.URL);err==nil{sourceLead=resolveImageURL(meta(parseHTML(cleaned),"og:image","twitter:image"),base)}
 normalizeArticleDocument(&article,sourceLead)
 stages=append(stages,traceImageBlocks("normalized",article.Blocks)...)
 _=json.NewEncoder(w).Encode(map[string]any{"url":input.URL,"sourceMetadataImage":sourceLead,"selectedImage":article.ImageURL,"byline":article.Byline,"images":stages})
}
