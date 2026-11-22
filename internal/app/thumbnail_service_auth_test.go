package app

import (
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)

func TestNewsGoThumbnailMachineIdentityScope(t *testing.T) {
 classify:=strings.Repeat("a",64)
 thumb:=strings.Repeat("b",64)
 gate:=&SuiteGate{NewsGoClassifyToken:classify,NewsGoThumbnailToken:thumb}
 reached:=0
 protected:=gate.Protect(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  reached++
  w.WriteHeader(http.StatusNoContent)
 }))
 request:=func(method,path,authorization,origin string)int {
  r:=httptest.NewRequest(method,"http://localhost:8083"+path,nil)
  if authorization!=""{r.Header.Set("Authorization",authorization)}
  if origin!=""{r.Header.Set("Origin",origin)}
  w:=httptest.NewRecorder()
  protected.ServeHTTP(w,r)
  if w.Header().Get("Cache-Control")!="no-store" && w.Code==http.StatusNoContent{
   t.Error("machine response missing no-store")
  }
  return w.Code
 }
 if got:=request("GET","/api/thumbnail?url=https%3A%2F%2Fexample.com%2Fone&width=256","Service "+thumb,"");got!=204{t.Fatalf("valid image machine identity got HTTP %d",got)}
 if got:=request("POST","/api/classify","Service "+classify,"");got!=204{t.Fatalf("existing classify identity broke: HTTP %d",got)}
 for _,tc:=range []struct{method,path,authorization,origin string}{
  {"POST","/api/thumbnail","Service "+thumb,""},
  {"GET","/api/classify","Service "+thumb,""},
  {"POST","/api/classify","Service "+thumb,""},
  {"GET","/api/article-image","Service "+thumb,""},
  {"GET","/api/read","Service "+thumb,""},
  {"GET","/api/pdf","Service "+thumb,""},
  {"GET","/api/suite/identity","Service "+thumb,""},
  {"GET","/api/activity/ping","Service "+thumb,""},
  {"GET","/api/diagnostic/extract","Service "+thumb,""},
  {"GET","/api/unknown-future-route","Service "+thumb,""},
  {"GET","/api/thumbnail","Service "+classify,""},
  {"GET","/api/thumbnail","",""},
  {"GET","/api/thumbnail","Service "+strings.Repeat("c",64),""},
  {"GET","/api/thumbnail","Service "+thumb,"https://newsgo.themauris.org"},
 }{
  if got:=request(tc.method,tc.path,tc.authorization,tc.origin);got!=http.StatusUnauthorized{
   t.Errorf("%s %s unexpected authorization HTTP %d",tc.method,tc.path,got)
  }
 }
 if reached!=2{t.Errorf("expected exactly two machine-authorized calls, got %d",reached)}
}

func TestNewsGoThumbnailCredentialCannotReuseClassifyToken(t *testing.T) {
 token:=strings.Repeat("a",64)
 for _,invalid:=range []string{"",strings.Repeat("f",63),strings.Repeat("G",64),token}{
  gate:=&SuiteGate{NewsGoClassifyToken:token,NewsGoThumbnailToken:invalid}
  r:=httptest.NewRequest(http.MethodGet,"http://localhost:8083/api/thumbnail",nil)
  r.Header.Set("Authorization","Service "+invalid)
  if gate.authorizedNewsGoThumbnail(r){t.Errorf("invalid or reused token authorized")}
 }
}
