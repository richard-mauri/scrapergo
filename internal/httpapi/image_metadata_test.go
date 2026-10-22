package httpapi

import (
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
 "clean-article-reader/internal/article"
)

func TestImageMetadataHandlerPreservesOrderAndContract(t *testing.T) {
 upstream:=http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  _=json.NewEncoder(w).Encode(article.Article{URL:"https://example.org/report",ImageURL:"https://example.org/lead.jpg",Blocks:[]article.Block{{Kind:"img",URL:"https://example.org/lead.jpg"},{Kind:"img",URL:"https://example.org/second.jpg"}}})
 })
 r:=httptest.NewRequest(http.MethodPost,"/api/image-metadata",strings.NewReader("{}"))
 w:=httptest.NewRecorder()
 ImageMetadataHandler(upstream).ServeHTTP(w,r)
 if w.Code!=http.StatusOK {t.Fatalf("status=%d",w.Code)}
 var payload struct { ImageURL string `json:"imageUrl"`; Candidates []string `json:"imageCandidates"`; ImagePresent bool `json:"imagePresent"` }
 if err:=json.Unmarshal(w.Body.Bytes(),&payload);err!=nil {t.Fatal(err)}
 if !payload.ImagePresent || payload.ImageURL!="https://example.org/lead.jpg" || len(payload.Candidates)!=2 || payload.Candidates[1]!="https://example.org/second.jpg" {t.Fatalf("payload=%+v",payload)}
}

func TestImageMetadataHandlerMethodAndUpstreamError(t *testing.T) {
 called:=false
 upstream:=http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){called=true;w.WriteHeader(403);_,_=w.Write([]byte(`{"error":"blocked"}`))})
 h:=ImageMetadataHandler(upstream)
 w:=httptest.NewRecorder()
 h.ServeHTTP(w,httptest.NewRequest(http.MethodGet,"/api/image-metadata",nil))
 if w.Code!=405 || called || w.Header().Get("Allow")!="POST" {t.Fatalf("method behavior %d called=%v",w.Code,called)}
 w=httptest.NewRecorder()
 h.ServeHTTP(w,httptest.NewRequest(http.MethodPost,"/api/image-metadata",nil))
 if w.Code!=403 || strings.TrimSpace(w.Body.String())!=`{"error":"blocked"}` {t.Fatalf("upstream error %d: %s",w.Code,w.Body.String())}
}
