package app

import (
 "fmt"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
 "time"
)

func TestScraperGoSuiteGateIdentityAndAPI(t *testing.T){
 token:=strings.Repeat("a",64)
 code:=200
 role:="guest"
 user:="guest-001"
 expired:=false
 calls:=0
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  calls++
  if r.URL.Path!="/api/session/profile"||
   r.Header.Get("Authorization")!="Bearer "+token||
   r.Header.Get("Origin")!=scrapergoOrigin {
   t.Errorf("profile request must use the trusted origin and bearer")
  }
  if code!=200{w.WriteHeader(code);return}
  expiry:=time.Now().Add(time.Hour)
  if expired{expiry=time.Now().Add(-time.Hour)}
  fmt.Fprintf(w,`{"user":%q,"role":%q,"expiresAt":%q}`,user,role,expiry.UTC().Format(time.RFC3339))
 }))
 defer server.Close()
 gate:=&SuiteGate{ProfileURL:server.URL+"/api/session/profile",Client:server.Client()}
 reached:=0
 handler:=gate.Protect(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){reached++;w.WriteHeader(200)}))
 request:=func(path,authorization string)int{
  req:=httptest.NewRequest("GET","http://localhost:8083"+path,nil)
  if authorization!=""{req.Header.Set("Authorization",authorization)}
  req.Header.Set("X-ScraperGo-Role","admin")
  w:=httptest.NewRecorder();handler.ServeHTTP(w,req)
  if strings.HasPrefix(path,"/api")&&w.Header().Get("Cache-Control")!="no-store"{
   t.Errorf("protected response %q missing no-store",path)
  }
  return w.Code
 }
 for _,path:=range []string{"/api","/api/read","/api/pdf","/api/classify","/api/image-metadata",
  "/api/article-image","/api/thumbnail","/api/diagnostic/extract",
  "/api/diagnostic/raw-html","/api/diagnostic/image-trace",
  "/api/diagnostic/classification-trace","/api/activity/ping",
  "/api/activity/identity","/api/unknown-future-endpoint"}{
  if got:=request(path,"");got!=401{t.Errorf("anonymous %s got %d",path,got)}
  if got:=request(path,"Bearer not-a-bearer");got!=401{t.Errorf("invalid bearer %s got %d",path,got)}
  if got:=request(path,"Bearer "+token);got!=200{t.Errorf("valid Guest %s got %d",path,got)}
 }
 if got:=request("/api/suite/identity","Bearer "+token);got!=200{t.Errorf("Guest identity got %d",got)}
 role="user";user="reader"
 if got:=request("/api/read","Bearer "+token);got!=200{t.Errorf("personal account got %d",got)}
 role="admin";user="operator"
 if got:=request("/api/diagnostic/extract","Bearer "+token);got!=200{t.Errorf("admin account got %d",got)}
 role="invalid"
 if got:=request("/api/read","Bearer "+token);got!=401{t.Errorf("spoofed role got %d",got)}
 role="user";user=""
 if got:=request("/api/read","Bearer "+token);got!=401{t.Errorf("blank user got %d",got)}
 user="reader";expired=true
 if got:=request("/api/read","Bearer "+token);got!=401{t.Errorf("expired account got %d",got)}
 expired=false;code=401
 if got:=request("/api/read","Bearer "+token);got!=401{t.Errorf("revoked account got %d",got)}
 code=503
 if got:=request("/api/read","Bearer "+token);got!=503{t.Errorf("Homepage outage got %d",got)}
 code=200
 for _,path:=range []string{"/","/tanner.jpg","/healthz","/anything-not-api"}{
  if got:=request(path,"");got!=200{t.Errorf("static bootstrap path %s got %d",path,got)}
 }
 if calls==0||reached==0{t.Fatal("verification and router not exercised")}
}

func TestScraperGoSuiteGateFailClosedOnNetworkError(t *testing.T){
 gate:=&SuiteGate{ProfileURL:"http://127.0.0.1:1/api/session/profile",Client:&http.Client{Timeout:time.Second}}
 handler:=gate.Protect(http.HandlerFunc(func(http.ResponseWriter,*http.Request){t.Fatal("unverified request reached app")}))
 req:=httptest.NewRequest("POST","http://localhost:8083/api/read",nil)
 req.Header.Set("Authorization","Bearer "+strings.Repeat("b",64))
 response:=httptest.NewRecorder()
 handler.ServeHTTP(response,req)
 if response.Code!=503{t.Fatalf("unexpected outage status: %d",response.Code)}
}
