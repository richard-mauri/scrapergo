package httpapi

import (
 "context"
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "net/url"
 "strings"
 "testing"
)

func TestClassificationPreservesPublisherStatus(t *testing.T) {
 for _,tc:=range []struct{name,body,classification string;want int}{
  {"denied",`{"kind":"publisher_access_denied","publisherHttpStatus":403}`,"unavailable",403},
  {"unauthorized",`{"kind":"publisher_access_denied","publisherHttpStatus":401}`,"unavailable",401},
  {"rate limited",`{"kind":"access_error","publisherHttpStatus":429}`,"unavailable",429},
  {"network error",`{"kind":"access_error"}`,"unavailable",0},
  {"success",`{"title":"Synthetic article","imageUrl":""}`,"article",0},
 } {
  t.Run(tc.name,func(t *testing.T) {
   status:=422
   if tc.classification=="article" {status=200}
   h:=&Classifier{
    ValidateURL:func(_ context.Context,s string)(*url.URL,error){return url.Parse(s)},
    Read:http.HandlerFunc(func(w http.ResponseWriter,_ *http.Request){w.WriteHeader(status);_,_=w.Write([]byte(tc.body))}),
   }
   address:="https://fixture.example.test/status-"+tc.name
   body,_:=json.Marshal(map[string]string{"url":address})
   w:=httptest.NewRecorder()
   h.ServeHTTP(w,httptest.NewRequest("POST","/api/classify",strings.NewReader(string(body))))
   var result classificationResult
   if err:=json.Unmarshal(w.Body.Bytes(),&result);err!=nil {t.Fatal(err)}
   if result.Classification!=tc.classification || result.PublisherHTTPStatus!=tc.want {
    t.Fatalf("classification result: %+v",result)
   }
  })
 }
}
