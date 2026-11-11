// Command duplicateprovenance summarizes privacy-minimized candidate duplicate
// counts at each extraction and reader stage. It never prints URLs, identifiers
// or article text.
package main

import (
 "bufio"
 "encoding/json"
 "flag"
 "fmt"
 "io"
 "os"
 "sort"

 "clean-article-reader/internal/diagnostics"
 "clean-article-reader/internal/extraction"
)

type totals struct {
 reached int
 beforeExact, afterExact int
 beforeOverlap, afterOverlap int
 addedExact, removedExact int
 addedOverlap, removedOverlap int
}
func hasBaseline(row diagnostics.RecoveryRecord) bool {
 for _,s:=range row.Stages {
  if s.Stage=="readability-render-text" && s.Outcome=="observed" {return true}
 }
 return false
}
func lastTraces(r io.Reader, n int) ([]diagnostics.RecoveryRecord,error) {
 scan:=bufio.NewScanner(r)
 scan.Buffer(make([]byte,4096),4<<20)
 records:=make([]diagnostics.RecoveryRecord,0,n)
 line:=0
 for scan.Scan() {
  line++
  var row diagnostics.RecoveryRecord
  if err:=json.Unmarshal(scan.Bytes(),&row);err!=nil {return nil,fmt.Errorf("invalid trace record at line %d: %w",line,err)}
  if !hasBaseline(row) {continue} // Ignore old records without candidate counts.
  if len(records)==n {copy(records,records[1:]);records[len(records)-1]=row} else {records=append(records,row)}
 }
 return records,scan.Err()
}
func compare(records []diagnostics.RecoveryRecord) map[string]*totals {
 out:=map[string]*totals{}
 for _,row:=range records {
  for _,stage:=range row.Stages {
   if stage.Outcome=="not_reached" || stage.Outcome=="" {continue}
   a:=out[stage.Stage]
   if a==nil {a=&totals{};out[stage.Stage]=a}
   a.reached++
   a.beforeExact+=stage.BeforeExactPairs
   a.afterExact+=stage.AfterExactPairs
   a.beforeOverlap+=stage.BeforeOverlapPairs
   a.afterOverlap+=stage.AfterOverlapPairs
   if d:=stage.AfterExactPairs-stage.BeforeExactPairs;d>0 {a.addedExact+=d} else {a.removedExact-=d}
   if d:=stage.AfterOverlapPairs-stage.BeforeOverlapPairs;d>0 {a.addedOverlap+=d} else {a.removedOverlap-=d}
  }
 }
 return out
}
func writeReport(w io.Writer, records []diagnostics.RecoveryRecord) {
 fmt.Fprintf(w,"Duplicate provenance: %d recent traced article reads (only new-format records)\n",len(records))
 fmt.Fprintln(w,"Stage                         Reads  Exact before/after  Exact added/removed  Overlap before/after  Overlap added/removed")
 stages:=compare(records)
 seen:=map[string]bool{}
 printStage:=func(name string) {
  v:=stages[name]
  if v==nil{return}
  seen[name]=true
  fmt.Fprintf(w,"%-29s %5d  %6d / %-6d  %6d / %-6d  %6d / %-6d     %6d / %-6d\n",
   name,v.reached,v.beforeExact,v.afterExact,v.addedExact,v.removedExact,
   v.beforeOverlap,v.afterOverlap,v.addedOverlap,v.removedOverlap)
 }
 for _,name:=range extraction.RecoveryStages {printStage(name)}
 var extra []string
 for name:=range stages {if !seen[name] {extra=append(extra,name)}}
 sort.Strings(extra)
 for _,name:=range extra {printStage(name)}
 fmt.Fprintln(w,"Counts are screening candidates, not proof of an error; an increase localizes a stage to investigate.")
 fmt.Fprintln(w,"The batch may include extra requests from browsers or other clients. Compare collection windows carefully.")
}
func main() {
 path:=flag.String("path","diagnostics/recovery.jsonl","private recovery trace JSONL path")
 last:=flag.Int("last",100,"most recent N trace records with duplicate metrics (1–2000)")
 flag.Parse()
 if *last<1 || *last>2000 || flag.NArg()!=0 {
  fmt.Fprintln(os.Stderr,"Usage: go run ./cmd/duplicateprovenance [-path diagnostics/recovery.jsonl] [-last 100]")
  os.Exit(2)
 }
 f,err:=os.Open(*path)
 if err!=nil {fmt.Fprintln(os.Stderr,"Cannot read local diagnostic data:",err);os.Exit(1)}
 defer f.Close()
 records,err:=lastTraces(f,*last)
 if err!=nil {fmt.Fprintln(os.Stderr,err);os.Exit(1)}
 if len(records)==0 {fmt.Fprintln(os.Stderr,"No new-format duplicate metrics found; build and run the PR #120 branch, then read articles.");os.Exit(1)}
 writeReport(os.Stdout,records)
}
