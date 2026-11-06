// Command recoveryreport summarizes local persistent recovery observations
// without printing article identifiers or URLs.
package main

import (
 "bufio"
 "encoding/json"
 "fmt"
 "os"
 "path/filepath"
 "sort"

 "clean-article-reader/internal/diagnostics"
)

type stageSummary struct { Reached, Changed, Fallback int }
func main() {
 path:=filepath.Join("diagnostics","recovery.jsonl")
 if len(os.Args)>2 { fmt.Fprintln(os.Stderr,"Usage: go run ./cmd/recoveryreport [path-to-recovery.jsonl]"); os.Exit(2) }
 if len(os.Args)==2 { path=os.Args[1] }
 f,err:=os.Open(path)
 if err!=nil { fmt.Fprintln(os.Stderr,"Cannot read recovery diagnostics:",err); os.Exit(1) }
 defer f.Close()
 scan:=bufio.NewScanner(f)
 scan.Buffer(make([]byte,4096),4<<20)
 total:=0
 unique:=make(map[string]struct{})
 publishers:=make(map[string]struct{})
 stages:=make(map[string]*stageSummary)
 for scan.Scan() {
  var row diagnostics.RecoveryRecord
  if err:=json.Unmarshal(scan.Bytes(),&row);err!=nil{
   fmt.Fprintln(os.Stderr,"Invalid recovery record at line",total+1,":",err);os.Exit(1)
  }
  total++
  unique[row.ArticleID]=struct{}{}
  publishers[row.Publisher]=struct{}{}
  for _,stage:=range row.Stages {
   x:=stages[stage.Stage]
   if x==nil { x=&stageSummary{};stages[stage.Stage]=x }
   switch stage.Outcome {
   case "modified": x.Reached++;x.Changed++
   case "fallback_selected": x.Reached++;x.Fallback++
   case "executed_unchanged": x.Reached++
   }
  }
 }
 if err:=scan.Err();err!=nil{fmt.Fprintln(os.Stderr,"Cannot scan diagnostics:",err);os.Exit(1)}
 fmt.Printf("Recovery trace: %d visits, %d distinct articles, %d publishers\n",total,len(unique),len(publishers))
 fmt.Println("Stage                          Reached  Changed  Fallback")
 names:=make([]string,0,len(stages))
 for name:=range stages {names=append(names,name)}
 sort.Strings(names)
 for _,name:=range names { x:=stages[name];fmt.Printf("%-30s %7d %8d %9d\n",name,x.Reached,x.Changed,x.Fallback) }
 fmt.Println("An executed change is not proof the change survived later stages or was beneficial.")
}
