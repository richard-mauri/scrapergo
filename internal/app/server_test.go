package app

import (
 "net/http"
 "clean-article-reader/internal/activity"
 "net/http/httptest"
 "strings"
 "testing"
)

func testHandlers() Handlers {
 reject:=http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){w.WriteHeader(http.StatusMethodNotAllowed)})
 return Handlers{
  Read:reject, Classify:reject, ImageMetadata:reject,
  ArticleImage:reject, Thumbnail:reject, DiagnosticExtract:reject,
  DiagnosticImageTrace:reject, TannerPhoto:[]byte{0xff,0xd8,0xff},
 }
}
func TestRouterPreservesRoutesAndCSP(t *testing.T){
 h:=Router(testHandlers(),Config{Port:"8083",Version:"0.2.85",Build:"Architecture PR #47"})
 for _,p:=range []string{"/api/read","/api/classify","/api/image-metadata","/api/article-image","/api/thumbnail","/api/diagnostic/extract","/api/diagnostic/image-trace"} {
  rec:=httptest.NewRecorder();h.ServeHTTP(rec,httptest.NewRequest(http.MethodGet,p,nil))
  if rec.Code!=http.StatusMethodNotAllowed {t.Errorf("%s route changed: %d",p,rec.Code)}
 }
 rec:=httptest.NewRecorder();h.ServeHTTP(rec,httptest.NewRequest(http.MethodGet,"/",nil))
 if rec.Code!=http.StatusOK || !strings.Contains(rec.Body.String(),"Architecture PR #47"){t.Fatalf("root page changed: %d",rec.Code)}
 if !strings.Contains(rec.Header().Get("Content-Security-Policy"),"img-src 'self' https: data:"){t.Fatal("image CSP changed")}
 rec=httptest.NewRecorder();h.ServeHTTP(rec,httptest.NewRequest(http.MethodGet,"/missing",nil))
 if rec.Code!=http.StatusNotFound {t.Fatalf("unknown path: %d",rec.Code)}
}
func TestTannerRoute(t *testing.T){
 h:=Router(testHandlers(),Config{})
 rec:=httptest.NewRecorder();h.ServeHTTP(rec,httptest.NewRequest(http.MethodGet,"/tanner.jpg",nil))
 if rec.Code!=http.StatusOK||rec.Header().Get("Content-Type")!="image/jpeg" {t.Fatalf("image route: %d",rec.Code)}
 rec=httptest.NewRecorder();h.ServeHTTP(rec,httptest.NewRequest(http.MethodPost,"/tanner.jpg",nil))
 if rec.Code!=http.StatusMethodNotAllowed||rec.Header().Get("Allow")!="GET, HEAD" {t.Fatalf("image method contract: %d",rec.Code)}
}

func TestOptionalActivityDoesNotChangeReaderRouting(t *testing.T){
 key:=strings.Repeat("k",40)
 a,err:=activity.NewActivityReporter("http://127.0.0.1:8788/api/admin/activity/ingest",key)
 if err!=nil{t.Fatal(err)}
 config:=Config{Port:"8088",Version:"0.2.85",Build:"Activity Test",Activity:a}
 handler:=Router(testHandlers(),config)
 root:=httptest.NewRecorder()
 handler.ServeHTTP(root,httptest.NewRequest("GET","http://127.0.0.1:8088/",nil))
 if root.Code!=200||!strings.Contains(root.Body.String(),"/api/activity/ping"){t.Fatalf("opt-in UI missing: %d",root.Code)}
 if strings.Contains(root.Body.String(),key){t.Fatal("private collector key exposed")}
 if !strings.Contains(root.Header().Get("Content-Security-Policy"),"script-src 'unsafe-inline'"){t.Fatal("reader CSP changed")}
 req:=httptest.NewRequest("POST","http://127.0.0.1:8088/api/activity/ping",strings.NewReader(`{"id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","visible":true,"interacted":true}`))
 req.Header.Set("Origin","http://127.0.0.1:8088");req.Header.Set("Content-Type","application/json")
 pong:=httptest.NewRecorder();handler.ServeHTTP(pong,req)
 if pong.Code!=204{t.Fatalf("ping status %d",pong.Code)}
 read:=httptest.NewRecorder();handler.ServeHTTP(read,httptest.NewRequest("GET","http://127.0.0.1:8088/api/read",nil))
 if read.Code!=405{t.Fatalf("read contract changed: %d",read.Code)}
 if got:=a.InFlight();got!=0{t.Fatalf("in-flight leaked: %d",got)}
 plain:=Router(testHandlers(),Config{Version:"0.2.85"})
 noActivity:=httptest.NewRecorder();plain.ServeHTTP(noActivity,httptest.NewRequest("GET","http://127.0.0.1/",nil))
 if strings.Contains(noActivity.Body.String(),"scrapergo-activity-tab-v1"){t.Fatal("telemetry active when unconfigured")}
}
