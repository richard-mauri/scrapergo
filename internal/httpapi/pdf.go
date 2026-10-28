package httpapi

import (
 "context"
 "net/http"
 "net/url"
)

// PDFHandler serves bounded public PDF data through the existing validated
// publisher fetch boundary. The browser never needs to fetch publisher PDFs
// directly, so the reader preserves original document pagination and columns.
func PDFHandler(download func(context.Context,string)([]byte,string,error)) http.Handler {
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.Method != http.MethodGet {
   w.Header().Set("Allow","GET")
   http.Error(w,"GET required",http.StatusMethodNotAllowed)
   return
  }
  address := r.URL.Query().Get("url")
  u,err := url.Parse(address)
  if err!=nil || u==nil || (u.Scheme!="https" && u.Scheme!="http") || u.User!=nil || u.Hostname()=="" {
   http.Error(w,"Invalid PDF URL",http.StatusBadRequest)
   return
  }
  data,_,err := download(r.Context(),address)
  if err!=nil {http.Error(w,"PDF unavailable",http.StatusUnprocessableEntity);return}
  w.Header().Set("Content-Type","application/pdf")
  w.Header().Set("Content-Disposition",`inline; filename="article.pdf"`)
  w.Header().Set("X-Content-Type-Options","nosniff")
  w.Header().Set("Cache-Control","no-store")
  w.Header().Set("Content-Security-Policy","frame-ancestors 'self'")
  _,_ = w.Write(data)
 })
}
