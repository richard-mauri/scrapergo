package httpapi

import (
 "context"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)

func TestReadHandlerPrefersContextAwareExtractor(t *testing.T) {
 called:=false
 h:=ReadHandler(ReadDependencies{
  Resolve:func(ctx context.Context,s string)(string,error){return s,nil},
  Download:func(ctx context.Context,s string)(string,string,error){return "<html>viewer</html>",s,nil},
  Hostname:func(s string)string{return "journal.example.org"},
  PublisherAccessMessage:func(error)(string,bool){return "",false},
  WriteExtracted:func(http.ResponseWriter,string,string){t.Fatal("old writer called")},
  WriteExtractedWithContext:func(ctx context.Context,w http.ResponseWriter,raw,source string){
   called=true
   if ctx==nil || source!="https://journal.example.org/article/view/37/33" {t.Fatalf("source=%q",source)}
   _,_ = w.Write([]byte(`{"title":"PDF article"}`))
  },
 })
 w:=httptest.NewRecorder()
 h.ServeHTTP(w,httptest.NewRequest(http.MethodPost,"/api/read",strings.NewReader(`{"url":"https://journal.example.org/article/view/37/33"}`)))
 if !called || w.Code!=200 {t.Fatalf("called=%t status=%d body=%s",called,w.Code,w.Body.String())}
}
