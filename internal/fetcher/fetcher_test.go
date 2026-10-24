package fetcher

import (
 "context"
 "errors"
 "net/http"
 "net/url"
 "strings"
 "testing"
)

func TestUnsafeURLRejectedBeforeNetworkAccess(t *testing.T) {
 cases:=[]string{
  "file:///etc/passwd",
  "https://user:password@example.org/",
  "http://localhost/article",
  "http://test.localhost/page",
  "http://127.0.0.1/article",
  "http://192.168.1.5/article",
 }
 for _,raw:=range cases {
  t.Run(raw,func(t *testing.T){
   if _,err:=ValidatedURL(context.Background(),raw);err==nil {t.Fatalf("unsafe URL allowed: %s",raw)}
  })
 }
}

func TestRetryAndPublisherAccessContract(t *testing.T) {
 if !AlternateHeaderRetry(http.StatusForbidden) {t.Fatal("403 retry removed")}
 for _,code:=range []int{http.StatusUnauthorized,http.StatusTooManyRequests,http.StatusInternalServerError} {
  if AlternateHeaderRetry(code) {t.Fatalf("unexpected alternate request for HTTP %d",code)}
 }
 for _,code:=range []int{http.StatusUnauthorized,http.StatusForbidden} {
  message,restricted:=PublisherAccessMessage(&PublisherHTTPError{Host:"publisher.example",StatusCode:code})
  if !restricted || !strings.Contains(message,"Publisher access restricted") {t.Fatalf("HTTP %d not classified as access restriction",code)}
 }
 if _,restricted:=PublisherAccessMessage(&PublisherHTTPError{Host:"publisher.example",StatusCode:http.StatusPaymentRequired});restricted {
  t.Fatal("unexpected inference of subscription restriction")
 }
 if _,restricted:=PublisherAccessMessage(errors.New("network timeout"));restricted {t.Fatal("ordinary network error classified as publisher restriction")}
}

func TestRequestProfiles(t *testing.T) {
 for _,setup:=range []func(*http.Request){SetBrowserHeaders,SetReadeckLikeHeaders} {
  req,_:=http.NewRequest(http.MethodGet,"https://example.org/",nil)
  setup(req)
  if req.Header.Get("User-Agent")=="" || req.Header.Get("Accept")=="" || req.Header.Get("Accept-Language")=="" {
   t.Fatal("missing expected request profile headers")
  }
 }
}

func TestWordPressPublicPostValidation(t *testing.T) {
 original, _ := url.Parse("https://sailinganarchy.com/2026/10/22/ferraris-100-foot-flying-yellow-monster-gets-its-hull/")
 const slug = "ferraris-100-foot-flying-yellow-monster-gets-its-hull"
 good := `[{"slug":"`+slug+`","link":"https://sailinganarchy.com/2026/10/22/`+slug+`/","status":"publish","title":{"rendered":"Ferrari"},"content":{"rendered":"<p>Valid article text</p>","protected":false}}]`
 if html, ok := decodePublicWPPost([]byte(good), original, slug); !ok || !strings.Contains(html,"Valid article text") {
  t.Fatalf("public post rejected: %v",ok)
 }
 for name, data := range map[string]string{
  "invalid-json": "{",
  "no-match": "[]",
  "duplicate": good[:len(good)-1]+","+good[1:],
  "protected": strings.Replace(good, `"protected":false`,`"protected":true`,1),
  "draft": strings.Replace(good, `"publish"`,`"draft"`,1),
  "mismatched-link": strings.Replace(good, "https://sailinganarchy.com/2026/10/22/", "https://another.example/2026/10/22/",1),
  "mismatched-slug": strings.Replace(good, `"slug":"`+slug+`"`, `"slug":"wrong"`,1),
  "empty-body": strings.Replace(good, "<p>Valid article text</p>", "",1),
 } {
  t.Run(name,func(t *testing.T) {
   if _,ok:=decodePublicWPPost([]byte(data),original,slug);ok { t.Fatal("untrusted response accepted") }
  })
 }
}
