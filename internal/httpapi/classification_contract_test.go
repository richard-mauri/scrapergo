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

// Frozen synthetic reader responses: no publisher URL or network request.
// These expectations intentionally describe the merged main baseline.
func TestClassificationStaticOutcomes(t *testing.T) {
 cases:=[]struct{name string;code int;body,classification,reason string; readable,image bool}{
  {"article",200,`{"title":"Synthetic report","blocks":[{"kind":"p","text":"A public paragraph."}],"imageUrl":""}`,"article","",true,false},
  {"article image",200,`{"title":"Illustrated report","blocks":[{"kind":"p","text":"Public text."}],"imageUrl":"https://images.example.test/example.jpg"}`,"article","",true,true},
  {"homepage",422,`{"kind":"page_index"}`,"homepage","homepage_or_index",false,false},
  {"unreadable",422,`{"kind":"extraction_error"}`,"unreadable","no_readable_article",false,false},
  {"publisher denied",422,`{"kind":"publisher_access_denied"}`,"unavailable","publisher_unavailable",false,false},
  {"fetch error",422,`{"kind":"access_error"}`,"unavailable","publisher_unavailable",false,false},
  {"malformed success",200,`{bad json`,"unknown","extraction_unavailable",false,false},
  {"unexpected error",500,`{"kind":"unexpected"}`,"unknown","extraction_unavailable",false,false},
 }
 for i,tc:=range cases {
  t.Run(tc.name,func(t *testing.T){
   address:="https://fixture.example.test/story-"+string(rune('a'+i))
   reads:=0
   handler:=&Classifier{
    ValidateURL:func(_ context.Context,raw string)(*url.URL,error){return url.Parse(raw)},
    Read:http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){reads++;w.WriteHeader(tc.code);_,_=w.Write([]byte(tc.body))}),
   }
   input,_:=json.Marshal(map[string]string{"url":address})
   rec:=httptest.NewRecorder()
   handler.ServeHTTP(rec,httptest.NewRequest(http.MethodPost,"/api/classify",strings.NewReader(string(input))))
   if rec.Code!=200 {t.Fatalf("http status %d: %s",rec.Code,rec.Body.String())}
   var got classificationResult
   if err:=json.Unmarshal(rec.Body.Bytes(),&got);err!=nil {t.Fatal(err)}
   if got.Classification!=tc.classification || got.Reason!=tc.reason || got.ReadableArticle!=tc.readable || got.ImageAvailable!=tc.image {
    t.Fatalf("got %+v; wanted classification=%s reason=%s readable=%v image=%v",got,tc.classification,tc.reason,tc.readable,tc.image)
   }
   if reads!=1 {t.Fatalf("reader calls=%d, want 1",reads)}
  })
 }
}

func TestClassificationInvalidMethodAndInput(t *testing.T) {
 h:=&Classifier{ValidateURL:func(_ context.Context,s string)(*url.URL,error){return nil,context.Canceled}}
 for _,tc:=range []struct{method,body string;want int}{
  {"GET","",405},
  {"POST","not json",400},
  {"POST",`{"url":""}`,400},
  {"POST",`{"url":"https://fixture.example.test/story"}`,422},
 }{
  w:=httptest.NewRecorder()
  h.ServeHTTP(w,httptest.NewRequest(tc.method,"/api/classify",strings.NewReader(tc.body)))
  if w.Code!=tc.want {t.Errorf("%s %q: got HTTP %d want %d",tc.method,tc.body,w.Code,tc.want)}
 }
}
