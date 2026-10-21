package main

import (
 "strings"
 "testing"
)

func TestReaderUsesNewsGoPalette(t *testing.T) {
 page:=renderedIndexHTML()
 for _, token:=range []string{"color-scheme:light dark","--panel:light-dark(#ffffff,#19232b)","--bg:light-dark(#f5f5f0,#0e141a)","--accent:light-dark(#126877,#69c6d3)","@media print","ScraperGo"} {
  if !strings.Contains(page,token){t.Errorf("reader HTML missing %q",token)}
 }
}
