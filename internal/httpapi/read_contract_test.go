package httpapi

import (
 "context"
 "errors"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)

func TestReadHandlerResolutionFailureDoesNotDownload(t *testing.T) {
 downloaded := false
 h := ReadHandler(ReadDependencies{
  Resolve: func(context.Context,string)(string,error){return "",errors.New("unsafe URL")},
  Download: func(context.Context,string)(string,string,error){downloaded=true;return "","",nil},
 })
 w := httptest.NewRecorder()
 h.ServeHTTP(w,httptest.NewRequest(http.MethodPost,"/api/read",strings.NewReader(`{"url":"http://127.0.0.1/private"}`)))
 if downloaded || w.Code != http.StatusUnprocessableEntity || !strings.Contains(w.Body.String(),"unsafe URL") {
  t.Fatalf("resolution failure: status=%d downloaded=%v body=%q",w.Code,downloaded,w.Body.String())
 }
 if w.Header().Get("Cache-Control")!="no-store"{t.Fatal("missing no-store header")}
}

func TestReadHandlerBodyLimitRejectsOversizeInput(t *testing.T) {
 called:=false
 h:=ReadHandler(ReadDependencies{
  Resolve:func(context.Context,string)(string,error){called=true;return "",nil},
 })
 w:=httptest.NewRecorder()
 body:=`{"url":"`+strings.Repeat("a",4096)+`"}`
 h.ServeHTTP(w,httptest.NewRequest(http.MethodPost,"/api/read",strings.NewReader(body)))
 if w.Code!=http.StatusBadRequest || called {
  t.Fatalf("oversize body: status=%d called=%v body=%q",w.Code,called,w.Body.String())
 }
}
