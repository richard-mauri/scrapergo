package app

import (
 "net/http"
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
