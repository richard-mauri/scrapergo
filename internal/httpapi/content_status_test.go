package httpapi

import "testing"

func TestClassifyContentStatus(t *testing.T) {
 cases := []struct{ kind, note, want, reason string }{
  {"restricted","","unknown",""},
  {"registration","","unknown",""},
  {"paywall","","unknown",""},
  {"","The publisher marks this article as access-restricted. The text below may be incomplete.","unknown",""},
  {"","","unknown",""},
  {"","ordinary article","unknown",""},
 }
 for _, tc := range cases {
  got, why := classifyContentStatus(tc.kind,tc.note)
  if got!=tc.want || why!=tc.reason {t.Fatalf("%q: got %q %q, want %q %q",tc.kind,got,why,tc.want,tc.reason)}
 }
}
