package ui

import (
 "strings"
 "testing"
)

func TestMandatorySuiteAuthAlwaysPresent(t *testing.T){
 html:=Render("0.0.0","Build 000001")
 for _,want:=range []string{
  "suite-auth-ready",
  "scrapergo-suite-handoff-v1",
  "function setAccess(ok)",
  "async function checkIdentity()",
  "scrapergoSuiteAuth",
  "if(input)input.disabled=!ok",
  "if(read)read.disabled=!ok",
  "const rejectedLocal=parsed.get('suite_mode')==='local'",
  "Homepage sign-in required",
  "Go to Homepage",
  "SuiteBannerStatus",
 }{
  if !strings.Contains(html,want){t.Errorf("mandatory browser bootstrap missing %q",want)}
 }
 if strings.Contains(html,"__SUITE_AUTH_JS__")||strings.Contains(html,"__SUITE_BANNER_SCRIPT__"){
  t.Fatal("unresolved browser auth placeholder")
 }
 if strings.Contains(html,"browser-only mode") || strings.Contains(html,"Browser only"){
  t.Fatal("anonymous browser identity must not be displayed as authorized access")
 }
}
func TestScraperGoReaderAttachesBearerForProtectedMedia(t *testing.T){
 for _,want:=range []string{
  "window.scrapergoSuiteAuth.fetch('/api/read'",
  "window.scrapergoSuiteAuth.isVerified()",
  "protectedObjectURL('/api/pdf?url='",
  "protectedObjectURL('/api/article-image?url='",
  "window.scrapergoSuiteAuth.whenReady().then",
  "void window.scrapergoSuiteAuth.whenReady().then(()=>renderBrowserArticle(article))",
  "function renderBrowserArticle(a)",
 }{
  if !strings.Contains(readerJS,want){t.Errorf("protected reader contract missing %q",want)}
 }
 for _,bad:=range []string{
  "await fetch('/api/read'",
  "frame.src='/api/pdf?url='",
  "img.src='/api/article-image?url='",
 }{
  if strings.Contains(readerJS,bad){t.Errorf("unauthenticated resource path %q",bad)}
 }
}
