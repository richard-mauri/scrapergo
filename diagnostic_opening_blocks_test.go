package main

import (
 articlemodel "clean-article-reader/internal/article"
 "clean-article-reader/internal/diagnostics"
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "bytes"
 "testing"
)

func TestOpeningBlockTracePreservesPartsAndLinks(t *testing.T) {
 blocks:=[]articlemodel.Block{{Kind:"p",Text:"Ruth Sherlock",Parts:[]articlemodel.TextPart{{Text:"Ruth Sherlock",URL:"https://www.npr.org/people/123"}}},{Kind:"p",Text:","},{Kind:"img",Role:"author",URL:"https://example.org/portrait.jpg"}}
 got:=diagnostics.TraceOpeningBlocks("readability",blocks)
 if len(got)!=3||got[0]["stage"]!="readability"||got[0]["index"]!=0 {t.Fatalf("trace missing blocks: %+v",got)}
 data,err:=json.Marshal(got);if err!=nil{t.Fatal(err)}
 var output []map[string]any
 if err:=json.Unmarshal(data,&output);err!=nil{t.Fatal(err)}
 parts,ok:=output[0]["parts"].([]any)
 if !ok||len(parts)!=1{t.Fatalf("inline link parts absent: %s",data)}
 if output[2]["role"]!="author"{t.Fatalf("image role missing: %s",data)}
}
func TestDiagnosticImageTraceRejectsNonLocalPeerForOpeningBlocks(t *testing.T) {
 req:=httptest.NewRequest(http.MethodPost,"/api/diagnostic/image-trace",bytes.NewBufferString(`{"url":"https://example.org/story"}`))
 req.RemoteAddr="203.0.113.20:41230"
 rec:=httptest.NewRecorder()
 (diagnostics.ImageTraceHandler{Extract:extractWithEvidence}).ServeHTTP(rec,req)
 if rec.Code!=http.StatusForbidden{t.Fatalf("expected localhost restriction, got %d",rec.Code)}
}
