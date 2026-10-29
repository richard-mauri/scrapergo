package bootstrap

import (
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)

func TestClassificationTraceOnlyLoopback(t *testing.T) {
 h:=classificationTraceHandler(http.HandlerFunc(func(http.ResponseWriter,*http.Request){}))
 tests:=[]struct{name,remote,method string; status int}{
  {"remoteDenied","192.0.2.12:60000","POST",http.StatusForbidden},
  {"hostHeaderCannotBypass","192.0.2.12:60000","POST",http.StatusForbidden},
  {"missingPayload","127.0.0.1:3456","POST",http.StatusBadRequest},
  {"methodDenied","127.0.0.1:3456","GET",http.StatusMethodNotAllowed},
 }
 for _,tt:=range tests {t.Run(tt.name,func(t *testing.T) {
  req:=httptest.NewRequest(tt.method,"http://localhost/api/diagnostic/classification-trace",strings.NewReader("{}"))
  req.RemoteAddr=tt.remote
  req.Host="localhost"
  recorder:=httptest.NewRecorder()
  h.ServeHTTP(recorder,req)
  if recorder.Code!=tt.status {t.Fatalf("status=%d, want %d: %s",recorder.Code,tt.status,recorder.Body.String())}
 })}
}
