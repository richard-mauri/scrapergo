package fetcher

import (
 "context"
 "testing"
)

func TestDownloadPDFRejectsUnsafeURLBeforeFetching(t *testing.T) {
 for _,raw:=range []string{"file:///etc/passwd","http://127.0.0.1/a.pdf","http://localhost/a.pdf","https://user:pass@example.org/a.pdf"} {
  if _,_,err:=DownloadPDF(context.Background(),raw);err==nil { t.Fatalf("accepted unsafe URL %s",raw) }
 }
}
