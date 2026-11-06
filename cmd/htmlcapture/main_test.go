package main

import (
 "context"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "os"
 "path/filepath"
 "strings"
 "testing"
 "time"
)

func TestCaptureWritesPrivateHTMLAndMetadata(t *testing.T) {
 folder:=t.TempDir()
 output:=filepath.Join(folder,"publisher.html")
 raw:="<html><body><article><h2>Update title</h2></article></body></html>"
 timestamp:=time.Date(2026,11,6,14,0,0,0,time.UTC)
 called:=0
 download:=func(_ context.Context,address string)(string,string,error){
  called++
  if address!="https://example.org/redirect" {t.Fatalf("URL %s",address)}
  return raw,"https://example.org/final",nil
 }
 info,err:=capture(context.Background(),"https://example.org/redirect",output,download,func()time.Time{return timestamp})
 if err!=nil{t.Fatal(err)}
 if called!=1 || info.FinalURL!="https://example.org/final" || info.RetrievedAt!="2026-11-06T14:00:00Z" {t.Fatalf("metadata: %+v",info)}
 want:=sha256.Sum256([]byte(raw))
 if info.SHA256!=hex.EncodeToString(want[:]) || info.Bytes!=len(raw){t.Fatalf("digest mismatch: %+v",info)}
 content,err:=os.ReadFile(output);if err!=nil || string(content)!=raw{t.Fatalf("HTML mismatch: %v",err)}
 sidecar,err:=os.ReadFile(output+".json");if err!=nil{t.Fatal(err)}
 var persisted captureMetadata
 if err:=json.Unmarshal(sidecar,&persisted);err!=nil{t.Fatal(err)}
 if persisted!=info{t.Fatalf("metadata mismatch: %+v vs %+v",persisted,info)}
 for _,path:=range []string{output,output+".json"} {
  st,err:=os.Stat(path);if err!=nil{t.Fatal(err)}
  if st.Mode().Perm()&0077!=0{t.Fatalf("not private: %s: %o",path,st.Mode().Perm())}
 }
}

func TestCaptureRefusesOverwriteWithoutNetwork(t *testing.T) {
 for _,existing:=range []string{"publisher.html","publisher.html.json"} {
  t.Run(existing,func(t *testing.T){
   folder:=t.TempDir();output:=filepath.Join(folder,"publisher.html")
   if err:=os.WriteFile(filepath.Join(folder,existing),[]byte("keep"),0600);err!=nil{t.Fatal(err)}
   called:=false
   _,err:=capture(context.Background(),"https://example.org/article",output,func(context.Context,string)(string,string,error){called=true;return "<html/>","",nil},time.Now)
   if err==nil || called{t.Fatalf("expected preflight refusal without fetch: %v",err)}
   b,err:=os.ReadFile(filepath.Join(folder,existing));if err!=nil || string(b)!="keep"{t.Fatalf("existing file overwritten: %v",err)}
  })
 }
}

func TestCaptureNoArtifactsOnFetchError(t *testing.T) {
 output:=filepath.Join(t.TempDir(),"publisher.html")
 _,err:=capture(context.Background(),"https://example.org",output,func(context.Context,string)(string,string,error){
  return "","",errors.New("access restricted")
 },time.Now)
 if err==nil || !strings.Contains(err.Error(),"access restricted"){t.Fatalf("unexpected error: %v",err)}
 for _,name:=range []string{output,output+".json"}{
  if _,err:=os.Stat(name);!errors.Is(err,os.ErrNotExist){t.Fatalf("unexpected file %s: %v",name,err)}
 }
}

func TestCaptureRejectsInvalidOutputWithoutNetwork(t *testing.T) {
 called:=false
 _,err:=capture(context.Background(),"https://example.org",filepath.Join(t.TempDir(),"capture.txt"),func(context.Context,string)(string,string,error){
  called=true;return "","",nil
 },time.Now)
 if err==nil || called{t.Fatalf("invalid output should prevent network: %v",err)}
}

func TestCaptureRejectsOversizedResult(t *testing.T) {
 output:=filepath.Join(t.TempDir(),"capture.html")
 _,err:=capture(context.Background(),"https://example.org",output,func(context.Context,string)(string,string,error){
  return strings.Repeat("x",5<<20+1),"https://example.org",nil
 },time.Now)
 if err==nil{t.Fatal("oversized response accepted")}
}
