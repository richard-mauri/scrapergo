package httpapi

import "testing"

func TestClassifyContentStatus(t *testing.T) {
 cases := []struct{ kind, note, want, reason string }{
  {"restricted","","partial","publisher_access_restricted"},
  {"registration","","partial","publisher_access_restricted"},
  {"paywall","","partial","publisher_access_restricted"},
  {"","The publisher marks this article as access-restricted. The text below may be incomplete.","partial","publisher_access_restricted"},
  {"","","unknown",""},
  {"","ordinary article","unknown",""},
 }
 for _, tc := range cases {
  got, why := classifyContentStatus(tc.kind,tc.note)
  if got!=tc.want || why!=tc.reason {t.Fatalf("%q: got %q %q, want %q %q",tc.kind,got,why,tc.want,tc.reason)}
 }
}
