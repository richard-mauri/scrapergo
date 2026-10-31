package httpapi

import (
 "context"
 "errors"
 "io"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)

func TestRawHTMLDiagnosticLocalAndExact(t *testing.T) {
 raw := "<!doctype html>\n<html><body><p>sample &amp; text</p></body></html>\n"
 calls := 0
 handler := RawHTMLDiagnosticHandler(func(_ context.Context, u string) (string,string,error) {
  calls++
  if u != "https://example.com/story" { t.Fatalf("unexpected url: %s",u) }
  return raw,"https://example.com/story",nil
 })
 req:=httptest.NewRequest(http.MethodPost,"http://localhost/api/diagnostic/raw-html",strings.NewReader(`{"url":"https://example.com/story"}`))
 req.RemoteAddr="127.0.0.1:32100"
 response:=httptest.NewRecorder()
 handler.ServeHTTP(response,req)
 if response.Code != 200 || response.Body.String()!=raw || calls != 1 {
  t.Fatalf("status=%d body=%q calls=%d",response.Code,response.Body.String(),calls)
 }
 if !strings.Contains(response.Header().Get("Content-Disposition"),"attachment") ||
  response.Header().Get("Cache-Control")!="no-store" ||
  response.Header().Get("Content-Security-Policy")!="sandbox" {t.Fatal("unsafe response headers")}
}

func TestRawHTMLDiagnosticRejectsRemoteWithoutDownload(t *testing.T) {
 calls:=0
 handler:=RawHTMLDiagnosticHandler(func(context.Context,string)(string,string,error){calls++;return "","",nil})
 for _,address:=range []string{"192.0.2.1:1000","203.0.113.2:1000",""} {
  req:=httptest.NewRequest(http.MethodPost,"/api/diagnostic/raw-html",strings.NewReader(`{"url":"https://example.com"}`))
  req.RemoteAddr=address
  w:=httptest.NewRecorder();handler.ServeHTTP(w,req)
  if w.Code!=403 {t.Fatalf("addr=%q status=%d",address,w.Code)}
 }
 if calls!=0 {t.Fatal("remote request invoked publisher downloader")}
}

func TestRawHTMLDiagnosticErrors(t *testing.T) {
 handler:=RawHTMLDiagnosticHandler(func(context.Context,string)(string,string,error){return "","",errors.New("blocked")})
 cases:=[]struct{method,body string;status int}{
  {"GET","",405},
  {"POST",`{"url":""}`,400},
  {"POST",`{"url":"file:///etc/passwd"}`,400},
  {"POST",`{"url":"https://example.org"}`,422},
 }
 for _,tc:=range cases {
  req:=httptest.NewRequest(tc.method,"/api/diagnostic/raw-html",strings.NewReader(tc.body))
  req.RemoteAddr="127.0.0.1:42"
  w:=httptest.NewRecorder();handler.ServeHTTP(w,req)
  if w.Code!=tc.status {t.Fatalf("%s body=%q got=%d expected=%d",tc.method,tc.body,w.Code,tc.status)}
  _,_=io.Copy(io.Discard,w.Body)
 }
}
