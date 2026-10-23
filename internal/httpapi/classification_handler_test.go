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

func TestClassificationUsesReaderAndURLValidation(t *testing.T) {
 calls:=0
 validate:=func(_ context.Context,raw string)(*url.URL,error) {
  if strings.Contains(raw,"private") {return nil,errors.New("blocked")}
  return url.Parse(raw)
 }
 read:=http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
  calls++
  if r.Method!=http.MethodPost {t.Errorf("reader method %s",r.Method)}
  w.Header().Set("Content-Type","application/json")
  w.Write([]byte("{\"title\":\"News\",\"imageUrl\":\"https://cdn.example.test/image.png\"}"))
 })
 handler:=&Classifier{Read:read,ValidateURL:validate}
 for _,tc:=range []struct{address string;status int;classification string}{
  {"https://news.example.test/p1",200,"article"},
  {"https://news.example.test/p1",200,"article"},
  {"https://private.test/item",422,""},
 } {
  rec:=httptest.NewRecorder()
  req:=httptest.NewRequest(http.MethodPost,"/api/classify",strings.NewReader("{\"url\":\""+tc.address+"\"}"))
  handler.ServeHTTP(rec,req)
  if rec.Code!=tc.status {t.Fatalf("%s HTTP %d, want %d",tc.address,rec.Code,tc.status)}
  if tc.classification!="" {
   var result map[string]any
   if err:=json.Unmarshal(rec.Body.Bytes(),&result);err!=nil{t.Fatal(err)}
   if result["classification"]!=tc.classification || result["readableArticle"]!=true || result["imageAvailable"]!=true {t.Fatalf("result: %+v",result)}
  }
 }
 if calls!=1 {t.Fatalf("cache failed or invalid URL bypassed validation; calls=%d",calls)}
}
