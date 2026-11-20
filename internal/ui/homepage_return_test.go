package ui

import (
 "strings"
 "testing"
)

func TestReaderHasSameTabHomepageReturn(t *testing.T) {
 page:=Render("0.2.85","test")
 if !strings.Contains(page,`class="suite-banner__home suite-type-control" href="https://home.themauris.org/"`){
  t.Fatal("reader header missing same-tab Homepage return")
 }
 if strings.Contains(page,`class="suite-banner__home suite-type-control" href="https://home.themauris.org/" target="_blank"`){
  t.Fatal("Homepage return must not open a new tab")
 }
 if !strings.Contains(page,".suite-banner__home:hover{text-decoration:underline}") {t.Fatal("shared link hover style missing")}

}
