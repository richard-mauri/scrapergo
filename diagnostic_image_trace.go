package main

import (
 "clean-article-reader/internal/structured"
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
 Gallery string `json:"gallery,omitempty"`
 Ancestors []string `json:"ancestors,omitempty"`
}
func traceImageBlocks(stage string, blocks []Block) []traceImage {
 out:=[]traceImage{}
 for i,b:=range blocks {
  if b.Kind!="img" {continue}
  out=append(out,traceImage{Stage:stage,Index:i,URL:b.URL,Role:b.Role,Caption:b.Text,SourceIdentity:imageSourceIdentity(b.URL),Gallery:b.Gallery})
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
  ancestry:=[]string{}
  for ancestor:=n.Parent;ancestor!=nil && len(ancestry)<8;ancestor=ancestor.Parent {
   ancestry=append(ancestry,ancestor.Tag+"#"+ancestor.Attr["id"]+"."+ancestor.Attr["class"])
  }
  out=append(out,traceImage{Stage:"source-dom",Index:len(out),URL:src,Caption:n.Attr["alt"],Parent:parent,ParentClass:pc,Grandparent:gparent,GrandparentClass:gc,SourceIdentity:imageSourceIdentity(src),Ancestors:ancestry})
  if len(out)>=150 {break}
 }
 return out
}

// traceOpeningBlocks records the extraction representation, including inline
// link parts that are not represented in Block.Links. Limit output by count.
func traceOpeningBlocks(stage string, blocks []Block) []map[string]any {
 out:=make([]map[string]any,0)
 limit:=len(blocks);if limit>45 {limit=45}
 for i:=0;i<limit;i++ {
  b:=blocks[i]
  out=append(out,map[string]any{"stage":stage,"index":i,"kind":b.Kind,"text":b.Text,"role":b.Role,"gallery":b.Gallery,"url":b.URL,"links":b.Links,"parts":b.Parts})
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
 cleaned:=structured.StripEmbeddedMedia(input.HTML)
 extracted,err:=extractWithEvidence(cleaned,input.URL)
 article:=extracted.Article
 if err!=nil{http.Error(w,"extraction failed",422);return}
 stages:=[]traceImage{}
 opening:=[]map[string]any{}
 opening=append(opening,traceOpeningBlocks("readability",article.Blocks)...)
 stages=append(stages,sourceImageTrace(cleaned,input.URL)...)
 stages=append(stages,traceImageBlocks("readability",article.Blocks)...)
 // Run the exact same stages in the same order as /api/read. Keep
 // historical phase snapshots for existing diagnostic consumers and capture
 // each named stage to expose portrait insertion/removal explicitly.
 for _,stage:=range articleAssemblyStages() {
  stage.Apply(&article,extracted.Evidence,cleaned)
  label:=stage.Name()
  stages=append(stages,traceImageBlocks(label,article.Blocks)...)
  opening=append(opening,traceOpeningBlocks(label,article.Blocks)...)
  switch label {
  case "recover-author-portrait":
   stages=append(stages,traceImageBlocks("recovered",article.Blocks)...)
   opening=append(opening,traceOpeningBlocks("recovered",article.Blocks)...)
  case "classify-source-images":
   stages=append(stages,traceImageBlocks("classified",article.Blocks)...)
   opening=append(opening,traceOpeningBlocks("classified",article.Blocks)...)
  case "normalize":
   stages=append(stages,traceImageBlocks("normalized",article.Blocks)...)
   opening=append(opening,traceOpeningBlocks("normalized",article.Blocks)...)
  }
 }
 sourceLead:=extracted.Evidence.MetadataImageURL
 _=json.NewEncoder(w).Encode(map[string]any{"url":input.URL,"sourceMetadataImage":sourceLead,"selectedImage":article.ImageURL,"byline":article.Byline,"images":stages,"openingBlocks":opening})
}
