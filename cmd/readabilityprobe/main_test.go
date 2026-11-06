package main

import (
 "bytes"
 "encoding/json"
 "os"
 "path/filepath"
 "strings"
 "testing"
)

func TestProbeSyntheticFixtures(t *testing.T) {
 cases:=[]struct{name string; headings, times int}{
  {"semantic-live.html", 5, 3},
  {"flattened-live.html", 1, 0},
  {"ordinary-article.html", 3, 1},
 }
 for _,tc:=range cases {
  t.Run(tc.name,func(t *testing.T){
   input:=filepath.Join("..","..","testdata","readability",tc.name)
   out:=t.TempDir()
   var stdout bytes.Buffer
   if err:=run(input,"https://example.org/test",out,&stdout);err!=nil {t.Fatal(err)}
   var r report
   if err:=json.Unmarshal(stdout.Bytes(),&r);err!=nil {t.Fatal(err)}
   if len(r.Source.Headings)!=tc.headings || len(r.Source.Times)!=tc.times {
    t.Fatalf("source structural metrics: %+v",r.Source)
   }
   for _,name:=range []string{"source.html","readability.html","readability.txt","report.json"}{
    b,err:=os.ReadFile(filepath.Join(out,name));if err!=nil || len(b)==0 {t.Fatalf("%s: %v",name,err)}
   }
  })
 }
}
func TestProbeRejectsUnsupportedBase(t *testing.T){
 var w bytes.Buffer
 input:=filepath.Join("..","..","testdata","readability","ordinary-article.html")
 if err:=run(input,"file:///etc/passwd","",&w);err==nil || !strings.Contains(err.Error(),"HTTP(S)") {t.Fatalf("expected URL rejection, got %v",err)}
}
func TestInspectCountsSemanticElements(t *testing.T){
 src:=`<main><article><header><time datetime="2026-11-06">11m ago</time><h2>Shipping <em>update</em></h2></header></article><section><h3>Other</h3></section></main>`
 snap,err:=inspect(strings.NewReader(src));if err!=nil{t.Fatal(err)}
 if snap.Articles!=1 || snap.Sections!=1 || len(snap.Headings)!=2 || len(snap.Times)!=1 || snap.Headings[0].Text!="Shipping update" || snap.Times[0].DateTime!="2026-11-06"{
  t.Fatalf("unexpected DOM summary: %+v",snap)
 }
}
