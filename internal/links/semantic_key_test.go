package links

import "testing"

func TestSemanticKeyPreservesSourceTextMatching(t *testing.T) {
 cases := []struct{a,b string; same bool}{
  {"The  editor's report", "the editor s   report",true},
  {"Alpha—beta", "alpha beta",true},
  {"Report 2026","Report 2025",false},
  {"Résumé of changes","résumé of changes",true},
 }
 for _, tc:=range cases {
  got:=SemanticKey(tc.a)==SemanticKey(tc.b)
  if got!=tc.same {t.Errorf("SemanticKey(%q) versus (%q): equal=%v, want %v",tc.a,tc.b,got,tc.same)}
 }
}
