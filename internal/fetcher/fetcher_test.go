package fetcher

import (
 "context"
 "errors"
 "net/http"
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
