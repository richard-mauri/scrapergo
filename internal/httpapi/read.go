package httpapi

import (
 "context"
 "encoding/json"
 "log"
 "net/http"
 "strings"
 "clean-article-reader/internal/accesspolicy"
)

// ReadDependencies supplies the application's existing URL resolution,
// secured retrieval and extraction without importing the executable package.
type ReadDependencies struct {
 Resolve func(context.Context,string)(string,error)
 Download func(context.Context,string)(string,string,error)
 PublisherAccessMessage func(error)(string,bool)
 Hostname func(string)string
 WriteExtracted func(http.ResponseWriter,string,string)
}

// ReadHandler owns the request and response contract. Access policy, URL
// validation, retrieval and article assembly remain injected services.
func ReadHandler(d ReadDependencies) http.Handler {
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
  if r.Method!="POST" {
   http.Error(w,"POST required",405)
   return
  }
  r.Body=http.MaxBytesReader(w,r.Body,4096)
  var in struct { URL string `json:"url"` }
  if err:=json.NewDecoder(r.Body).Decode(&in);err!=nil {
   http.Error(w,"Invalid request JSON",400)
   return
  }
  address,err:=d.Resolve(r.Context(),in.URL)
  if err!=nil {
   w.Header().Set("Content-Type","application/json; charset=utf-8")
   w.Header().Set("Cache-Control","no-store")
   w.WriteHeader(422)
   _=json.NewEncoder(w).Encode(map[string]string{"error":err.Error()})
   return
  }
  log.Printf("resolve publisher=%s google_news=%t",d.Hostname(address),strings.Contains(in.URL,"news.google.com/"))
  raw,finalURL,err:=d.Download(r.Context(),address)
  w.Header().Set("Content-Type","application/json; charset=utf-8")
  w.Header().Set("Cache-Control","no-store")
  if err!=nil {
   if message,denied:=d.PublisherAccessMessage(err);denied {
    w.WriteHeader(http.StatusUnprocessableEntity)
    _=json.NewEncoder(w).Encode(map[string]string{"error":message,"kind":"publisher_access_denied","url":address})
    return
   }
   w.WriteHeader(http.StatusUnprocessableEntity)
   _=json.NewEncoder(w).Encode(map[string]string{"error":err.Error(),"kind":"access_error"})
   return
  }
  if accesspolicy.ExplicitPreview(raw) { w.Header().Set("X-ScraperGo-Content-Status", "partial") }
  d.WriteExtracted(w,raw,finalURL)
 })
}
