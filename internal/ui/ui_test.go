package ui

import (
 "strings"
 "testing"
)

func TestRenderInjectsVersionAndEmbeddedAssets(t *testing.T) {
 page:=Render("0.2.85","Architecture PR #46")
 for _,required:=range []string{
  "<div class=\"suite-type-title\">ScraperGo</div>",
  "v0.2.85", "Build Architecture PR #46",
  "<style>","<script>","id=\"form\"",
 } {
  if !strings.Contains(page,required) {t.Errorf("render missing %q",required)}
 }
 for _,placeholder:=range []string{"__READER_CSS__","__READER_JS__","__SUITE_BANNER__"} {
  if strings.Contains(page,placeholder) {t.Errorf("unexpanded placeholder %s",placeholder)}
 }
 if !strings.Contains(Template(),"__SUITE_BANNER__") {t.Fatal("template must retain shared-banner slot")}
}

func TestEmbeddedAssetsRetainInlineCSPContract(t *testing.T) {
 page:=Render("0.2.85","test")
 if strings.Count(page,"<script>")!=1 || strings.Count(page,"</script>")!=1 {t.Fatal("script injection changed")}
 if strings.Count(page,"<style>")!=1 || strings.Count(page,"</style>")!=1 {t.Fatal("style injection changed")}
 if strings.Contains(page,"<script src=") || strings.Contains(page,"<link rel=\"stylesheet\"") {
  t.Fatal("inline CSP contract changed")
 }
}

func TestSharedAmbientBackgroundIsOptInAndPrintSafe(t *testing.T) {
 page:=Render("0.2.85","test")
 for _,want:=range []string{`<body class="suite-ambient-background">`, `.suite-ambient-background {`, `radial-gradient(`, `background-image: none`} {
  if !strings.Contains(page,want) {t.Errorf("missing shared ambient background contract: %s",want)}
 }
}

func TestScraperGoSharedBannerPreservesAccountDOM(t *testing.T) {
 page:=Render("0.2.85","test")
 for _,required:=range []string{
  `<header class="suite-ui suite-banner"`,
  `id="suite-auth-status"`,
  `class="suite-banner__account-label suite-auth-status"`,
  `aria-live="polite"`,
  "Browser-only mode",
  `src="/tanner.jpg"`,
  `.suite-banner__bottom`,
 } {
  if !strings.Contains(page,required) {t.Errorf("missing inline account state: %s",required)}
 }
 if strings.Count(page,`id="suite-auth-status"`)!=1 {t.Fatal("account status must have exactly one DOM element")}
 for _,obsolete:=range []string{"Build Build", "Local preferences", "border:1px solid var(--suite-line);border-radius:var(--suite-radius)"} {
  if strings.Contains(page,obsolete) {t.Errorf("obsolete shared banner presentation %q",obsolete)}
 }
 for _,want:=range []string{`suite-banner__portrait-caption">The Mauris`, `class="suite-type-title">ScraperGo`} {
  if !strings.Contains(page,want) {t.Errorf("missing refined banner %q",want)}
 }
 for _,obsolete:=range []string{"PERSONAL PORTAL", `suite-banner__family`, "border-bottom:1px solid var(--suite-line)","border-top:1px solid var(--suite-line)"} {
  if strings.Contains(page,obsolete) {t.Errorf("unwanted legacy banner decoration %q",obsolete)}
 }
}
