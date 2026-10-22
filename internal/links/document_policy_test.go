package links

import "testing"

func TestIsDocumentURLPreservesLinkEligibility(t *testing.T) {
 cases:=[]struct{url string; want bool}{
  {"https://example.org/story",true},
  {"http://example.org/another?ref=a",true},
  {"https://example.org/",true},
  {"https://example.org/report.pdf",false},
  {"https://example.org/photo.jpg?size=large",false},
  {"https://example.org/archive.zip",false},
  {"https://example.org/watch.mp4",false},
  {"javascript:alert(1)",false},
  {"https://user:password@example.org/story",false},
  {"/relative/story",false},
  {"not a url",false},
 }
 for _,tc:=range cases {
  if got:=IsDocumentURL(tc.url);got!=tc.want {t.Errorf("IsDocumentURL(%q)=%v want %v",tc.url,got,tc.want)}
 }
}
