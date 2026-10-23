package httpapi

import (
 "context"
 "errors"
 "net/http"
 "net/http/httptest"
 "net/url"
 "strings"
 "testing"
)

func TestThumbnailValidationPrecedesImageRetrieval(t *testing.T) {
 calls:=0
 handler:=NewThumbnailHandler(
  func(_ context.Context,raw string)(*url.URL,error){
   if strings.Contains(raw,"private") {return nil,errors.New("not public")}
   return url.Parse(raw)
  },
  http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){calls++;http.Error(w,"missing",http.StatusNotFound)}),
 )
 for _,tc:=range []struct{method,address string;status int}{
  {http.MethodPost,"/api/thumbnail",405},
  {http.MethodGet,"/api/thumbnail?url=https://news.example.test/p&width=1",400},
  {http.MethodGet,"/api/thumbnail?url=https://private.example.test/item",422},
  {http.MethodGet,"/api/thumbnail?url=https://news.example.test/p",404},
 }{
  rec:=httptest.NewRecorder()
  handler.ServeHTTP(rec,httptest.NewRequest(tc.method,tc.address,nil))
  if rec.Code!=tc.status {t.Fatalf("%s %s: got %d want %d",tc.method,tc.address,rec.Code,tc.status)}
 }
 if calls!=1 {t.Fatalf("image retrieval called %d times; wanted one",calls)}
}
