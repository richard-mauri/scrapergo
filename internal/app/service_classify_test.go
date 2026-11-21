package app

import (
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)

func TestNewsGoServiceIdentityOnlyAuthorizesClassification(t *testing.T) {
 key:=strings.Repeat("d",64)
 gate:=&SuiteGate{NewsGoClassifyToken:key}
 calls:=0
 app:=gate.Protect(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  calls++
  w.WriteHeader(http.StatusNoContent)
 }))
 check:=func(method,path,authorization,origin string,want int){
  t.Helper()
  req:=httptest.NewRequest(method,"http://127.0.0.1:8083"+path,nil)
  if authorization!=""{req.Header.Set("Authorization",authorization)}
  if origin!=""{req.Header.Set("Origin",origin)}
  rec:=httptest.NewRecorder()
  app.ServeHTTP(rec,req)
  if rec.Code!=want{t.Errorf("%s %s auth=%q origin=%q: got %d want %d",method,path,authorization[:min(len(authorization),8)],origin,rec.Code,want)}
 }
 valid:="Service "+key
 check("POST","/api/classify",valid,"",204)
 check("POST","/api/classify",valid,"https://evil.example",401)
 for _,path:=range []string{"/api/read","/api/pdf","/api/thumbnail","/api/suite/identity","/api/classify/other","/api/diagnostic/extract","/api/activity/ping"}{
  check("POST",path,valid,"",401)
 }
 check("GET","/api/classify",valid,"",401)
 check("POST","/api/classify","","",401)
 check("POST","/api/classify","Bearer "+key,"",401)
 check("POST","/api/classify","Service "+strings.Repeat("e",64),"",401)
 check("POST","/api/classify","Service "+strings.Repeat("D",64),"",401)
 gate.NewsGoClassifyToken=""
 check("POST","/api/classify",valid,"",401)
 gate.NewsGoClassifyToken="bad"
 check("POST","/api/classify",valid,"",401)
 if calls!=1{t.Fatalf("restricted service identity reached application %d times",calls)}
}
