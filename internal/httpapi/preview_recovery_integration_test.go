package httpapi

import (
 "context"
 "encoding/json"
 "errors"
 "net/http"
 "net/http/httptest"
 "net/url"
 "strings"
 "testing"
)

func TestCapturedPreviewContractAndDenialRecovery(t *testing.T) {
 const address = "https://example.org/article"
 // Synthetic editorial and preview content, not copyrighted publisher HTML.
 preview := `<!doctype html><html><body><article><h1>Sample report</h1><p>First public paragraph.</p><p>Second public paragraph.</p><p>You have a preview view of this article while we are checking your access. When we have confirmed access, the full article content will load.</p></article></body></html>`
 calls:=0
 read:=ReadHandler(ReadDependencies{
  Resolve:func(_ context.Context,u string)(string,error){return u,nil},
  Download:func(_ context.Context,u string)(string,string,error){
   calls++
   if calls==1{return "", "", errors.New("temporary denied")}
   return preview,u,nil
  },
  Hostname:func(string)string{return "example.org"},
  PublisherAccessMessage:func(err error)(string,bool){if err!=nil{return "denied",true};return "",false},
  WriteExtracted:func(w http.ResponseWriter,raw,source string){
   if !strings.Contains(raw,"First public paragraph.") || !strings.Contains(raw,"Second public paragraph.") {t.Fatal("lost public preview text")}
   _=json.NewEncoder(w).Encode(map[string]any{"title":"Sample report","imageUrl":"","blocks":[]map[string]string{{"kind":"p","text":"First public paragraph."},{"kind":"p","text":"Second public paragraph."}}})
  },
 })
 classifier:=&Classifier{
  Read:read,
  ValidateURL:func(_ context.Context,u string)(*url.URL,error){return url.Parse(u)},
 }
 classify:=func() classificationResult {
  t.Helper()
  w:=httptest.NewRecorder()
  classifier.ServeHTTP(w,httptest.NewRequest(http.MethodPost,"/api/classify",strings.NewReader(`{"url":"https://example.org/article"}`)))
  if w.Code!=200 {t.Fatalf("classification http=%d: %s",w.Code,w.Body.String())}
  var got classificationResult
  if err:=json.Unmarshal(w.Body.Bytes(),&got);err!=nil {t.Fatal(err)}
  return got
 }
 first:=classify()
 if first.Classification!="unavailable" || first.Reason!="publisher_access_denied" {t.Fatalf("first call: %+v",first)}
 second:=classify()
 if second.Classification!="article" || second.ContentStatus!="partial" || second.ContentReason!="publisher_explicit_preview" {t.Fatalf("recovered preview: %+v",second)}
 if calls!=2 {t.Fatalf("failure was cached, requests=%d",calls)}
}
