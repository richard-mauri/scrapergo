package ui

import (
 "strings"
 "testing"
)

func TestReaderUsesSystemSansSerif(t *testing.T) {
 for _, rule := range []string{
  `.paper h2{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;`,
  `.content{font:19px/1.75 -apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;`,
 } {
  if !strings.Contains(readerCSS, rule) {
   t.Fatalf("missing system sans-serif typography rule: %s", rule)
  }
 }
 if strings.Contains(readerCSS, "Georgia,serif") {
  t.Fatal("legacy serif reader font remains")
 }
}
