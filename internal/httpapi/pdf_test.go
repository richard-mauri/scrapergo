package httpapi

import (
 "context"
 "net/http/httptest"
 "strings"
 "testing"
)

func TestPDFHandlerResponse(t *testing.T) {
 calls:=0
 h:=PDFHandler(func(ctx context.Context,u string)([]byte,string,error){
  calls++
  if u!="https://journal.example.org/paper.pdf" {t.Fatalf("unexpected URL %q",u)}
  return []byte("%PDF-1.7 test"),u,nil
 })
 for _,raw:=range []string{"file:///etc/passwd","http://localhost@evil.example/paper.pdf",""}{
  w:=httptest.NewRecorder()
  h.ServeHTTP(w,httptest.NewRequest("GET","/api/pdf?url="+raw,nil))
  if w.Code!=400 {t.Fatalf("%s: %d",raw,w.Code)}
 }
 w:=httptest.NewRecorder()
 h.ServeHTTP(w,httptest.NewRequest("GET","/api/pdf?url=https%3A%2F%2Fjournal.example.org%2Fpaper.pdf",nil))
 if w.Code!=200 || w.Header().Get("Content-Type")!="application/pdf" || !strings.HasPrefix(w.Body.String(),"%PDF-") || calls!=1 {t.Fatalf("bad PDF response: %d, calls %d",w.Code,calls)}
}
