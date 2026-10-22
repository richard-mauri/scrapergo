package httpapi

import (
 "context"
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
