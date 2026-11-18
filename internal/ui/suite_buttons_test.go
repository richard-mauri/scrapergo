package ui

import (
 "strings"
 "testing"

 suiteui "github.com/richard-mauri/homepage/shared/ui"
)

func TestReaderUsesHomepageCanonicalButtons(t *testing.T) {
 css:=string(suiteui.ButtonCSS)
 if len(css)==0 {t.Fatal("Homepage embedded button CSS is empty")}
 page:=Template()
 if !strings.Contains(page,css) {t.Fatal("ScraperGo must inline Homepage's exact embedded button CSS")}
 for _,id:=range []string{"read","txt","md","print"} {
  if !strings.Contains(templateHTML,`class="ui-button" id="`+id+`"`) {t.Errorf("reader action %s does not use canonical component",id)}
 }
 for _,required:=range []string{"--text:var(--ink)","--line:var(--border)"} {
  if !strings.Contains(readerCSS,required) {t.Errorf("missing ScraperGo semantic token mapping %q",required)}
 }
 for _,old:=range []string{"button{font:inherit", "button:disabled{", "button:hover:not(:disabled){", ".toolbar button{", ".suite-signin-actions button{"} {
  if strings.Contains(readerCSS,old) {t.Errorf("obsolete local button component remains: %q",old)}
 }
 for _,required:=range []string{"button.ui-button", ":hover", ":focus-visible", ":disabled", "prefers-reduced-motion"} {
  if !strings.Contains(css,required) {t.Errorf("Homepage component missing %q",required)}
 }
 if strings.Contains(css,"@import") {t.Fatal("shared CSS must not add an external runtime dependency")}
}
