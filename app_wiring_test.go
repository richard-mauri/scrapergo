package main

import (
 "clean-article-reader/internal/bootstrap"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)

func TestCentralHandlerWiringPreservesReadMethodLimit(t *testing.T) {
 read := bootstrap.NewArticleReadHandler()
 metadata := bootstrap.NewArticleMetadataHandler(read)
 for _,tc:=range []struct{name string;handler http.Handler}{
  {"read",read},{"metadata",metadata},
 } {
  t.Run(tc.name,func(t *testing.T){
   recorder:=httptest.NewRecorder()
   tc.handler.ServeHTTP(recorder,httptest.NewRequest(http.MethodGet,"/api/"+tc.name,nil))
   if recorder.Code!=http.StatusMethodNotAllowed || !strings.Contains(recorder.Body.String(),"POST required") {
    t.Fatalf("GET response=%d %q",recorder.Code,recorder.Body.String())
   }
  })
 }
}
