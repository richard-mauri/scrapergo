package ui

import (
 "strings"
 "testing"
)

func TestReaderHasSameTabHomepageReturn(t *testing.T) {
 page:=Render("0.2.85","test")
 if !strings.Contains(page,`class="suite-home-link" href="https://home.themauris.org/"`){
  t.Fatal("reader header missing same-tab Homepage return")
 }
 if strings.Contains(page,`class="suite-home-link" target="_blank"`){
  t.Fatal("Homepage return must not open a new tab")
 }
 if !strings.Contains(page,".suite-home-link:hover{background:rgba(96,151,172,.16);color:var(--accent,#69c6d3);text-decoration:underline;opacity:1}") {t.Fatal("standardized hover border style missing")}

}
