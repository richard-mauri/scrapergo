package pdfarticle

import (
 "context"
 "errors"
 "strings"
 "testing"
)

func TestRejectsInvalidPDF(t *testing.T) {
 cases := []struct{name string; data []byte; want error}{
  {"empty",nil,ErrPDFTooLarge},
  {"too large",make([]byte,MaxPDFBytes+1),ErrPDFTooLarge},
  {"HTML",[]byte("<html>not a PDF</html>"),nil},
  {"broken PDF",[]byte("%PDF-1.7\\nnot valid"),nil},
 }
 for _,tc := range cases {t.Run(tc.name,func(t *testing.T){
  _,err := Extract(context.Background(),tc.data,"https://journal.example.org/paper.pdf")
  if err==nil {t.Fatal("expected error")}
  if tc.want!=nil && !errors.Is(err,tc.want) {t.Fatalf("got %v, want %v",err,tc.want)}
 })}
}
func TestInvalidSource(t *testing.T) {
 _,err := Extract(context.Background(),[]byte("%PDF-1.7\\n"),"file:///tmp/paper.pdf")
 if err==nil || !strings.Contains(err.Error(),"source") {t.Fatalf("got %v",err)}
}
