package activity

import (
 "context"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
 "time"
)

// A newly verified reader arrival is reported immediately, independent of
// the unchanged one-minute periodic activity ticker.
func TestVerifiedReaderArrivalReportsImmediately(t *testing.T) {
 reports:=make(chan struct{},2)
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
  if r.URL.Path=="/api/session/profile" {
   _,_=w.Write([]byte(`{"user":"guest-1234abcd","activitySession":"`+strings.Repeat("1",64)+`","activityUser":"`+strings.Repeat("2",64)+`"}`))
   return
  }
  if r.URL.Path=="/api/admin/activity/ingest" {
   if r.Header.Get("X-Activity-Service")!="scrapergo" {t.Error("wrong reporter")}
   select {case reports<-struct{}{}:default:}
   w.WriteHeader(http.StatusNoContent)
   return
  }
  http.NotFound(w,r)
 }))
 defer server.Close()
 reporter,err:=NewActivityReporter(server.URL+"/api/admin/activity/ingest",strings.Repeat("s",40))
 if err!=nil {t.Fatal(err)}
 ctx,cancel:=context.WithCancel(context.Background())
 defer cancel()
 go reporter.Run(ctx)
 request:=httptest.NewRequest(http.MethodPost,"http://127.0.0.1:8088/api/activity/ping",strings.NewReader(`{"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","visible":true}`))
 request.Header.Set("Origin","http://127.0.0.1:8088")
 request.Header.Set("Content-Type","application/json")
 request.Header.Set("Authorization","Bearer "+strings.Repeat("a",64))
 response:=httptest.NewRecorder()
 reporter.Ping(response,request)
 if response.Code!=http.StatusNoContent {t.Fatalf("ping %d",response.Code)}
 select {
 case <-reports:
 case <-time.After(3*time.Second):t.Fatal("verified arrival waited for periodic report")
 }
 if got:=reporter.snapshot();got.Authenticated!=1 || len(got.Users)!=1 || got.Users[0].User!="guest-1234abcd" {t.Fatalf("guest not attributed: %+v",got)}
}
