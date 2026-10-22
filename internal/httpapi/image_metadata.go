// Package httpapi owns HTTP presentation adapters, independent of retrieval
// and the application executable.
package httpapi

import (
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "strings"

 "clean-article-reader/internal/article"
)

// ImageMetadataHandler delegates article retrieval and access policy to the
// application's existing article handler. It does not fetch URLs independently.
func ImageMetadataHandler(readArticle http.Handler) http.Handler {
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
  w.Header().Set("Content-Type","application/json; charset=utf-8")
  w.Header().Set("Cache-Control","no-store")
  if r.Method!=http.MethodPost {
   w.Header().Set("Allow","POST")
   http.Error(w,"POST required",http.StatusMethodNotAllowed)
   return
  }
  recorder:=httptest.NewRecorder()
  readArticle.ServeHTTP(recorder,r)
  result:=recorder.Result()
  defer result.Body.Close()
  if result.StatusCode!=http.StatusOK {
   w.WriteHeader(result.StatusCode)
   var failure json.RawMessage
   if json.NewDecoder(result.Body).Decode(&failure)==nil && json.Valid(failure) {
    _,_=w.Write(failure)
   } else {
    _=json.NewEncoder(w).Encode(map[string]string{"error":"article lookup failed"})
   }
   return
  }
  var a article.Article
  if err:=json.NewDecoder(result.Body).Decode(&a);err!=nil {
   http.Error(w,"invalid extraction response",http.StatusInternalServerError)
   return
  }
  candidates:=ImageCandidates(a)
  selected:=""
  if len(candidates)>0 {selected=candidates[0]}
  _=json.NewEncoder(w).Encode(struct{
   URL string `json:"url"`
   ImageURL string `json:"imageUrl"`
   ImagePresent bool `json:"imagePresent"`
   Candidates []string `json:"imageCandidates"`
  }{URL:a.URL,ImageURL:selected,ImagePresent:selected!="",Candidates:candidates})
 })
}

// ImageCandidates retains the reader's hero-first and source-block order.
func ImageCandidates(a article.Article) []string {
 var out []string
 seen:=make(map[string]bool)
 add:=func(raw string) {
  raw=strings.TrimSpace(raw)
  if raw!="" && !seen[raw] {
   seen[raw]=true
   out=append(out,raw)
  }
 }
 add(a.ImageURL)
 for _,b:=range a.Blocks {if b.Kind=="img" {add(b.URL)}}
 return out
}
