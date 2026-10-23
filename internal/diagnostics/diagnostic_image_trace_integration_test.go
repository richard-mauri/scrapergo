package diagnostics_test

import (
 "clean-article-reader/internal/bootstrap"
 "clean-article-reader/internal/diagnostics"
 "bytes"
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)

func TestImageTraceRefusesRemoteClients(t *testing.T) {
 rec:=httptest.NewRecorder()
 req:=httptest.NewRequest(http.MethodPost,"/api/diagnostic/image-trace",strings.NewReader(`{}`))
 req.RemoteAddr="203.0.113.17:45000"
 (diagnostics.ImageTraceHandler{Extract:bootstrap.ExtractWithEvidence}).ServeHTTP(rec,req)
 if rec.Code!=http.StatusForbidden {t.Fatalf("expected forbidden, got %d",rec.Code)}
}
func TestImageTraceStagesAndSourceContainers(t *testing.T) {
 raw:=`<!doctype html><html><head><title>A sample report</title><meta property="og:image" content="https://example.org/other.jpg"></head><body><main><article><h1>A sample report</h1><div class="byline"><img src="https://example.org/portrait.jpg" alt="Author"/></div><p>`+strings.Repeat("The article contains documented facts and substantial reported narrative. ",15)+`</p><figure class="hero"><img src="https://example.org/first.jpg" alt="First scene"/><figcaption>First scene caption</figcaption></figure><p>`+strings.Repeat("The next paragraph contains additional context and analysis. ",15)+`</p></article></main></body></html>`
 payload,_:=json.Marshal(map[string]string{"url":"https://example.org/news","html":raw})
 req:=httptest.NewRequest(http.MethodPost,"/api/diagnostic/image-trace",bytes.NewReader(payload))
 req.RemoteAddr="127.0.0.1:52000"
 rec:=httptest.NewRecorder()
 (diagnostics.ImageTraceHandler{Extract:bootstrap.ExtractWithEvidence}).ServeHTTP(rec,req)
 if rec.Code!=http.StatusOK {t.Fatalf("status=%d body=%s",rec.Code,rec.Body.String())}
 var result struct{Images []diagnostics.TraceImage `json:"images"`}
 if err:=json.Unmarshal(rec.Body.Bytes(),&result);err!=nil{t.Fatal(err)}
 stages:=map[string]bool{}
 container:=false
 ancestry:=false
 for _,v:=range result.Images {stages[v.Stage]=true;if v.Stage=="source-dom"&&v.Parent=="figure"&&v.ParentClass=="hero"{container=true;if len(v.Ancestors)>0&&strings.Contains(v.Ancestors[0],"figure") {ancestry=true}}}
 for _,step:=range []string{"source-dom","readability","recovered","classified","normalized"}{if !stages[step]{t.Errorf("missing trace stage %s",step)}}
 if !container {t.Fatal("source image container context missing")}
 if !ancestry {t.Fatal("source image ancestor chain missing")}
}

func TestMissingSourceImageReportDistinguishesAbsentLeadFromRenditions(t *testing.T) {
 source:=[]diagnostics.TraceImage{
  {Stage:"source-dom",URL:"https://example.org/lead.jpg",SourceIdentity:"lead"},
  {Stage:"source-dom",URL:"https://example.org/inline-1024x768.jpg",SourceIdentity:"inline"},
  {Stage:"source-dom",URL:"https://example.org/inline.jpg",SourceIdentity:"inline"},
 }
 final:=[]diagnostics.TraceImage{{Stage:"assembled-final",URL:"https://example.org/inline.jpg",SourceIdentity:"inline"}}
 missing:=diagnostics.MissingSourceImageReport(source,final)
 if len(missing)!=1||missing[0].URL!="https://example.org/lead.jpg" {t.Fatalf("missing lead comparison: %+v",missing)}
}
