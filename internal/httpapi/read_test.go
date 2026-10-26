package httpapi

import (
 "context"
 "encoding/json"
 "errors"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)

func TestReadHandlerCallsServicesAndPreservesContract(t *testing.T) {
 var resolved,downloaded,written bool
 h:=ReadHandler(ReadDependencies{
  Resolve:func(_ context.Context,u string)(string,error){resolved=true;if u!="https://news.example.test/story"{t.Fatalf("url %q",u)};return u,nil},
  Download:func(_ context.Context,u string)(string,string,error){downloaded=true;return "<article>text</article>",u,nil},
  Hostname:func(string)string{return "news.example.test"},
  PublisherAccessMessage:func(error)(string,bool){return "",false},
  WriteExtracted:func(w http.ResponseWriter,raw,source string){written=true;_,_=w.Write([]byte(`{"title":"Example"}`))},
 })
 w:=httptest.NewRecorder()
 h.ServeHTTP(w,httptest.NewRequest(http.MethodPost,"/api/read",strings.NewReader(`{"url":"https://news.example.test/story"}`)))
 if w.Code!=200||!resolved||!downloaded||!written||w.Header().Get("Cache-Control")!="no-store"||!strings.Contains(w.Body.String(),"Example"){t.Fatalf("response=%d body=%q headers=%v",w.Code,w.Body.String(),w.Header())}
}
func TestReadHandlerPreservesErrorAndMethodHandling(t *testing.T) {
 fail:=errors.New("fetch denied")
 h:=ReadHandler(ReadDependencies{
  Resolve:func(_ context.Context,s string)(string,error){return s,nil},
  Download:func(context.Context,string)(string,string,error){return "","",fail},
  Hostname:func(string)string{return "example.test"},
  PublisherAccessMessage:func(error)(string,bool){return "Publisher access restricted",true},
 })
 for _,tc:=range []struct{method,body string;status int;kind string}{
  {"GET","",405,"POST required"},
  {"POST","bad json",400,"Invalid request JSON"},
  {"POST",`{"url":"https://example.test/story"}`,422,"publisher_access_denied"},
 } {
  w:=httptest.NewRecorder()
  h.ServeHTTP(w,httptest.NewRequest(tc.method,"/api/read",strings.NewReader(tc.body)))
  if w.Code!=tc.status||!strings.Contains(w.Body.String(),tc.kind){t.Fatalf("%s status=%d body=%q",tc.method,w.Code,w.Body.String())}
 }
}


type testPublisherStatusError struct { status int }
func (e testPublisherStatusError) Error() string { return "simulated publisher error" }
func (e testPublisherStatusError) PublisherStatusCode() int {return e.status}

func TestReadPublisherStatusMetadata(t *testing.T) {
 for _,tc:=range []struct{name string;err error;denied bool;wantStatus int;wantKind string}{
  {"forbidden",testPublisherStatusError{403},true,403,"publisher_access_denied"},
  {"unauthorized",testPublisherStatusError{401},true,401,"publisher_access_denied"},
  {"rate limited",testPublisherStatusError{429},false,429,"access_error"},
  {"transport failure",errors.New("dial failure"),false,0,"access_error"},
 } {
  t.Run(tc.name,func(t *testing.T){
   handler:=ReadHandler(ReadDependencies{
    Resolve:func(_ context.Context,s string)(string,error){return s,nil},
    Download:func(context.Context,string)(string,string,error){return "","",tc.err},
    PublisherAccessMessage:func(error)(string,bool){return "Publisher access restricted",tc.denied},
    Hostname:func(string)string{return "fixture.example.test"},
   })
   w:=httptest.NewRecorder()
   handler.ServeHTTP(w,httptest.NewRequest(http.MethodPost,"/api/read",strings.NewReader(`{"url":"https://fixture.example.test/story"}`)))
   if w.Code!=422 {t.Fatalf("status %d",w.Code)}
   var payload struct { Kind string `json:"kind"`; PublisherHTTPStatus int `json:"publisherHttpStatus"`; Error string `json:"error"` }
   if err:=json.Unmarshal(w.Body.Bytes(),&payload);err!=nil {t.Fatal(err)}
   if payload.Kind!=tc.wantKind || payload.PublisherHTTPStatus!=tc.wantStatus {t.Fatalf("response: %+v",payload)}
   if tc.denied && !strings.Contains(payload.Error,"Publisher HTTP") {t.Fatalf("no visible publisher status: %s",payload.Error)}
   if tc.wantStatus==0 && strings.Contains(w.Body.String(),"publisherHttpStatus") {t.Fatalf("invented status: %s",w.Body.String())}
  })
 }
}
