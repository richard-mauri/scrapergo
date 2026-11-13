package activity

import (
 "context"
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
 "time"
)

func TestOptionalReporterConfiguration(t *testing.T){
 if r,e:=NewActivityReporter("","");e!=nil||r!=nil{t.Fatalf("disabled: %v",e)}
 for _,url:=range []string{"http://public.example/api/admin/activity/ingest","http://127.0.0.1:8788/other","https://x:y@example.com/api/admin/activity/ingest"}{
  if _,e:=NewActivityReporter(url,strings.Repeat("k",40));e==nil{t.Errorf("accepted %q",url)}
 }
 if _,e:=NewActivityReporter("http://127.0.0.1:8788/api/admin/activity/ingest","weak");e==nil{t.Error("accepted short key")}
 if r,e:=NewActivityReporter("http://127.0.0.1:8788/api/admin/activity/ingest",strings.Repeat("k",40));e!=nil||r==nil{t.Fatalf("valid local endpoint: %v",e)}
}

func TestSameOriginAnonymousPresenceAndWork(t *testing.T){
 a,_:=NewActivityReporter("http://127.0.0.1:8788/api/admin/activity/ingest",strings.Repeat("k",40))
 now:=time.Now().UTC().Truncate(time.Second)
 a.now=func()time.Time{return now}
 ping:=func(origin string)int{
  req:=httptest.NewRequest("POST","http://127.0.0.1:8088/api/activity/ping",strings.NewReader(`{"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","visible":true,"interacted":true}`))
  req.Header.Set("Origin",origin);req.Header.Set("Content-Type","application/json")
  w:=httptest.NewRecorder();a.Ping(w,req);return w.Code
 }
 if got:=ping("https://hostile.example");got!=403{t.Errorf("cross origin: %d",got)}
 if got:=ping("http://127.0.0.1:8088");got!=204{t.Fatalf("same origin ping: %d",got)}
 p:=a.snapshot()
 if p.Active2!=1||p.Active5!=1||p.Active15!=1||p.Active60!=1||p.Anonymous!=1||p.Foreground!=1||p.LastInteraction==""{t.Fatalf("snapshot %+v",p)}
 done:=a.Working()
 if got:=a.snapshot().InFlight;got!=1{t.Fatalf("request work=%d",got)}
 done()
 if got:=a.snapshot().InFlight;got!=0{t.Fatalf("work leaked=%d",got)}
 now=now.Add(3*time.Minute)
 if p=a.snapshot();p.Active2!=0||p.Active5!=1{t.Fatalf("window rollover %+v",p)}
}

func TestScraperCollectorWireContract(t *testing.T){
 secret:=strings.Repeat("z",40)
 var report activityPayload
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.Header.Get("Authorization")!="Bearer "+secret||r.Header.Get("X-Activity-Service")!="scrapergo"{t.Errorf("server credentials missing")}
  if r.Header.Get("X-Activity-Timestamp")==""||r.Header.Get("X-Activity-Nonce")==""{t.Error("timestamp/replay fields missing")}
  if err:=json.NewDecoder(r.Body).Decode(&report);err!=nil{t.Error(err)}
  w.WriteHeader(204)
 }))
 defer server.Close()
 a,err:=NewActivityReporter(server.URL+"/api/admin/activity/ingest",secret)
 if err!=nil{t.Fatal(err)}
 a.mu.Lock();a.presence[strings.Repeat("b",32)]=activityPresence{seen:time.Now(),visible:true};a.mu.Unlock()
 if err:=a.report(context.Background());err!=nil{t.Fatal(err)}
 if report.Active2!=1||report.Anonymous!=1||report.Active60!=1||report.Authenticated!=0{t.Fatalf("aggregate %+v",report)}
}

func TestBrowserScriptKeepsServiceKeyPrivate(t *testing.T) {
 s:=BrowserScript()
 if !strings.Contains(s,"/api/activity/ping")||!strings.Contains(s,"sessionStorage"){t.Fatal("ephemeral same-origin script missing")}
 if strings.Contains(s,"HOMEPAGE_ACTIVITY_KEY_SCRAPERGO")||strings.Contains(s,"SCRAPERGO_ACTIVITY_URL"){t.Fatal("server-only config exposed")}
}
