package ui

import (
 "strings"
 "testing"
)

func TestRenderInjectsVersionAndEmbeddedAssets(t *testing.T) {
 page:=Render("0.2.85","Architecture PR #46")
 for _,required:=range []string{
  "ScraperGo · v0.2.85 · Architecture PR #46",
  "<style>","<script>","id=\"form\"",
 } {
  if !strings.Contains(page,required) {t.Errorf("render missing %q",required)}
 }
 for _,placeholder:=range []string{"__READER_CSS__","__READER_JS__","__APP_VERSION__"} {
  if strings.Contains(page,placeholder) {t.Errorf("unexpanded placeholder %s",placeholder)}
 }
 if !strings.Contains(Template(),"__APP_VERSION__") {t.Fatal("template must not hard-code version")}
}

func TestEmbeddedAssetsRetainInlineCSPContract(t *testing.T) {
 page:=Render("0.2.85","test")
 if strings.Count(page,"<script>")!=1 || strings.Count(page,"</script>")!=1 {t.Fatal("script injection changed")}
 if strings.Count(page,"<style>")!=1 || strings.Count(page,"</style>")!=1 {t.Fatal("style injection changed")}
 if strings.Contains(page,"<script src=") || strings.Contains(page,"<link rel=\"stylesheet\"") {
  t.Fatal("inline CSP contract changed")
 }
}

func TestScraperGoAccountStatusInlineWithHomepageLink(t *testing.T) {
 page:=Render("0.2.85","test")
 for _,required:=range []string{
  `<div class="suite-auth-line"><a class="suite-home-link"`,
  `<span id="suite-auth-status" class="suite-auth-status" role="status" aria-live="polite">`,
  "Guest — not signed in",
  ".suite-auth-line{display:flex",
  ".suite-auth-status{color:var(--accent);font:700",
 } {
  if !strings.Contains(page,required) {t.Errorf("missing inline account state: %s",required)}
 }
 if strings.Contains(page,`<p id="suite-auth-status"`) {t.Fatal("account status must not occupy a separate paragraph")}
}
