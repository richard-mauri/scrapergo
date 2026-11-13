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

func TestVerifiedHandoffCountsSessionsAndUsersWithoutExportingIDs(t *testing.T) {
 secret:=strings.Repeat("k",40)
 bearer:=strings.Repeat("a",64)
 sessionOne:=strings.Repeat("1",64)
 sessionTwo:=strings.Repeat("2",64)
 userTag:=strings.Repeat("3",64)
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
  if r.URL.Path!="/api/session/profile" || r.Header.Get("Origin")!="https://scrapergo.themauris.org" {
   t.Errorf("unexpected profile request: %s %s",r.URL.Path,r.Header.Get("Origin"))
  }
  if r.Header.Get("Authorization")!="Bearer "+bearer {w.WriteHeader(401);return}
  w.Header().Set("Content-Type","application/json")
  _,_=w.Write([]byte(`{"user":"richard","activitySession":"`+sessionOne+`","activityUser":"`+userTag+`"}`))
 }))
 defer server.Close()
 a,err:=NewActivityReporter(server.URL+"/api/admin/activity/ingest",secret)
 if err!=nil{t.Fatal(err)}
 req:=httptest.NewRequest("POST","http://127.0.0.1:8088/api/activity/ping",strings.NewReader(`{"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","visible":true}`))
 req.Header.Set("Origin","http://127.0.0.1:8088")
 req.Header.Set("Content-Type","application/json")
 req.Header.Set("Authorization","Bearer "+bearer)
 w:=httptest.NewRecorder();a.Ping(w,req)
 if w.Code!=204{t.Fatalf("verified ping: %d",w.Code)}
 now:=a.now()
 a.mu.Lock()
 a.presence["second-tab"]=activityPresence{seen:now,sessionKey:sessionOne,userKey:userTag,visible:true}
 a.presence["second-login"]=activityPresence{seen:now,sessionKey:sessionTwo,userKey:userTag,visible:true}
 a.presence["guest"]=activityPresence{seen:now,visible:true}
 a.mu.Unlock()
 p:=a.snapshot()
 if p.Active2!=3||p.Authenticated!=2||p.UniqueUsers!=1||p.Anonymous!=1 || len(p.Users)!=1 || p.Users[0].User!="richard" || p.Users[0].Sessions!=1 {
  t.Fatalf("expected two sessions, one user, one guest: %+v",p)
 }
 body,_:=json.Marshal(p)
 if strings.Contains(string(body),sessionOne)||strings.Contains(string(body),userTag) {
  t.Fatal("collector aggregate exposed identity tags")
 }
}

func TestRevokedAndMalformedScraperHandoffsAreGuestPresence(t *testing.T) {
 secret:=strings.Repeat("s",40)
 valid:=strings.Repeat("a",64)
 revoked:=false
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
  if revoked {w.WriteHeader(http.StatusUnauthorized);return}
  _,_=w.Write([]byte(`{"user":"richard","activitySession":"`+strings.Repeat("1",64)+`","activityUser":"`+strings.Repeat("2",64)+`"}`))
 }))
 defer server.Close()
 a,err:=NewActivityReporter(server.URL+"/api/admin/activity/ingest",secret)
 if err!=nil{t.Fatal(err)}
 now:=time.Now().UTC().Truncate(time.Second)
 a.now=func()time.Time{return now}
 ping:=func(token string)int{
  req:=httptest.NewRequest(http.MethodPost,"http://127.0.0.1:8088/api/activity/ping",
   strings.NewReader(`{"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","visible":true}`))
  req.Header.Set("Origin","http://127.0.0.1:8088")
  req.Header.Set("Content-Type","application/json")
  if token!=""{req.Header.Set("Authorization","Bearer "+token)}
  w:=httptest.NewRecorder();a.Ping(w,req);return w.Code
 }
 if got:=ping(valid);got!=204{t.Fatalf("valid ping %d",got)}
 if p:=a.snapshot();p.Authenticated!=1||p.UniqueUsers!=1||p.Anonymous!=0{t.Fatalf("valid profile %+v",p)}
 revoked=true;now=now.Add(16*time.Second)
 if got:=ping(valid);got!=204{t.Fatalf("revoked ping %d",got)}
 if p:=a.snapshot();p.Authenticated!=0||p.UniqueUsers!=0||p.Anonymous!=1{t.Fatalf("revoked bearer must become guest %+v",p)}
 now=now.Add(16*time.Second)
 if got:=ping("invalid");got!=204{t.Fatalf("malformed ping %d",got)}
 if p:=a.snapshot();p.Authenticated!=0||p.Anonymous!=1{t.Fatalf("malformed bearer must remain guest %+v",p)}
}

func TestBrowserHandoffPreservesReaderAndRemovesTokenFragment(t *testing.T) {
 script:=BrowserScript()
 for _,required:=range []string{
  "new URLSearchParams(location.hash.slice(1))",
  "hash.get('suite_token')",
  "hash.get('suite_mode')==='local'",
  "sessionStorage.removeItem(handoffKey)",
  "history.replaceState(null,'',location.pathname+location.search)",
  "headers.Authorization='Bearer '+bearer",
  "fetch('/api/activity/ping'",
 } {
  if !strings.Contains(script,required){t.Errorf("missing ScraperGo handoff invariant: %s",required)}
 }
 if strings.Contains(script,"?suite_token=") || strings.Contains(script,"/api/session/renew") {
  t.Fatal("browser handoff cannot use token query strings or renew Homepage login")
 }
}

func TestTemporaryHomepageProfileOutageDoesNotInventGuest(t *testing.T) {
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
  w.WriteHeader(http.StatusServiceUnavailable)
 }))
 defer server.Close()
 a,err:=NewActivityReporter(server.URL+"/api/admin/activity/ingest",strings.Repeat("k",40))
 if err!=nil{t.Fatal(err)}
 req:=httptest.NewRequest(http.MethodPost,"http://127.0.0.1:8088/api/activity/ping",
  strings.NewReader(`{"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","visible":true}`))
 req.Header.Set("Origin","http://127.0.0.1:8088")
 req.Header.Set("Content-Type","application/json")
 req.Header.Set("Authorization","Bearer "+strings.Repeat("a",64))
 w:=httptest.NewRecorder();a.Ping(w,req)
 if w.Code!=http.StatusServiceUnavailable{t.Fatalf("outage should not fabricate guest: %d",w.Code)}
 p:=a.snapshot()
 if p.Active2!=0||p.Anonymous!=0||p.Authenticated!=0{t.Fatalf("outage created false presence: %+v",p)}
}

func TestScraperGoIdentityStatusGuestSignedInAndUnavailable(t *testing.T) {
 valid:=strings.Repeat("a",64)
 unavailable:=false
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
  if unavailable {w.WriteHeader(http.StatusServiceUnavailable);return}
  if r.Header.Get("Authorization")!="Bearer "+valid {w.WriteHeader(http.StatusUnauthorized);return}
  _,_=w.Write([]byte(`{"user":"richard","activitySession":"`+strings.Repeat("1",64)+`","activityUser":"`+strings.Repeat("2",64)+`"}`))
 }))
 defer server.Close()
 a,err:=NewActivityReporter(server.URL+"/api/admin/activity/ingest",strings.Repeat("k",40))
 if err!=nil{t.Fatal(err)}
 check:=func(bearer string)(int,string){
  req:=httptest.NewRequest(http.MethodGet,"http://127.0.0.1:8088/api/activity/identity",nil)
  if bearer!=""{req.Header.Set("Authorization","Bearer "+bearer)}
  w:=httptest.NewRecorder();a.IdentityStatus(w,req)
  return w.Code,w.Body.String()
 }
 if code,body:=check("");code!=200||!strings.Contains(body,`"guest"`){t.Fatalf("guest status %d %s",code,body)}
 if code,body:=check(valid);code!=200||!strings.Contains(body,`"signed-in"`)||!strings.Contains(body,"richard"){
  t.Fatalf("authenticated status %d %s",code,body)
 }
 if code,body:=check(strings.Repeat("b",64));code!=200||!strings.Contains(body,`"guest"`){
  t.Fatalf("revoked status %d %s",code,body)
 }
 unavailable=true
 if code,_:=check(valid);code!=503{t.Fatalf("outage status %d",code)}
}

func TestSilentRecoveryDoesNotNavigateOrLeakParentCredential(t *testing.T) {
 script:=BrowserScript()
 for _,required:=range []string{
  "https://home.themauris.org/api/handoff/scrapergo-recover",
  "method:'POST',mode:'cors',credentials:'include'",
  "hadHandoff=!!bearer",
  "Session expired — Sign in again",
  "Unable to verify sign-in",
  "Signed in as ",
 } {
  if !strings.Contains(script,required){t.Errorf("missing silent recovery invariant %q",required)}
 }
 if strings.Contains(script,"location.assign(")||strings.Contains(script,"location.replace(")||
  strings.Contains(script,"window.location=")||strings.Contains(script,"/api/session/renew") {
  t.Fatal("silent recovery must not navigate or renew Homepage parent login")
 }
}

func TestDirectScraperGoVisitProbesExistingHomepageSession(t *testing.T) {
 script:=BrowserScript()
 for _,fragment:=range []string{
  "let localOnly=false",
  "localOnly=true",
  "if(localOnly) {statusElement.textContent='Guest — not signed in';return;}",
  "let identity=bearer?await identityResult():{state:'guest'}",
  "const recovery=await recoverHandoff()",
  "hadHandoff?'Session expired — Sign in again':'Guest — not signed in'",
  "hadHandoff=true",
 } {
  if !strings.Contains(script,fragment){t.Errorf("missing direct-login invariant: %q",fragment)}
 }
 if strings.Contains(script,"location.assign(")||strings.Contains(script,"location.replace(") {
  t.Fatal("direct login probe must never navigate")
 }
}
