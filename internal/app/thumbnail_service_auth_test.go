package app

import (
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)

// The existing machine credential is shared by the two explicitly authorized
// NewsGo operations, not by unrelated ScraperGo APIs.
func TestNewsGoSharedMachineIdentityScope(t *testing.T) {
 token:=strings.Repeat("a",64)
 gate:=&SuiteGate{NewsGoClassifyToken:token}
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
 if got:=request("GET","/api/thumbnail?url=https%3A%2F%2Fexample.com%2Fone&width=256","Service "+token,"");got!=204{t.Fatalf("valid thumbnail machine identity got HTTP %d",got)}
 if got:=request("POST","/api/classify","Service "+token,"");got!=204{t.Fatalf("valid classification machine identity got HTTP %d",got)}
 for _,tc:=range []struct{method,path,authorization,origin string}{
  {"POST","/api/thumbnail","Service "+token,""},
  {"GET","/api/classify","Service "+token,""},
  {"GET","/api/article-image","Service "+token,""},
  {"GET","/api/image-metadata","Service "+token,""},
  {"POST","/api/image-metadata","Service "+token,""},
  {"GET","/api/read","Service "+token,""},
  {"POST","/api/read","Service "+token,""},
  {"GET","/api/pdf","Service "+token,""},
  {"GET","/api/suite/identity","Service "+token,""},
  {"GET","/api/activity/ping","Service "+token,""},
  {"GET","/api/diagnostic/extract","Service "+token,""},
  {"GET","/api/unknown-future-route","Service "+token,""},
  {"GET","/api/thumbnail","",""},
  {"GET","/api/thumbnail","Service "+strings.Repeat("c",64),""},
  {"GET","/api/thumbnail","Service "+strings.Repeat("f",63),""},
  {"GET","/api/thumbnail","Service "+strings.Repeat("G",64),""},
  {"GET","/api/thumbnail","Service "+token,"https://newsgo.themauris.org"},
  {"POST","/api/classify","Service "+token,"https://newsgo.themauris.org"},
 }{
  if got:=request(tc.method,tc.path,tc.authorization,tc.origin);got!=http.StatusUnauthorized{
   t.Errorf("%s %s unexpected authorization HTTP %d",tc.method,tc.path,got)
  }
 }
 if reached!=2{t.Errorf("expected exactly two machine-authorized calls, got %d",reached)}
}

func TestNewsGoSharedMachineIdentityFailClosed(t *testing.T) {
 for _,invalid:=range []string{"",strings.Repeat("f",63),strings.Repeat("G",64)}{
  gate:=&SuiteGate{NewsGoClassifyToken:invalid}
  for _,tc:=range []struct{method,path string}{
   {http.MethodGet,"/api/thumbnail"},
   {http.MethodPost,"/api/classify"},
  }{
   r:=httptest.NewRequest(tc.method,"http://localhost:8083"+tc.path,nil)
   r.Header.Set("Authorization","Service "+invalid)
   if gate.authorizedNewsGoThumbnail(r) || gate.authorizedNewsGoClassification(r) {
    t.Errorf("invalid token authorized %s %s",tc.method,tc.path)
   }
  }
 }
}
