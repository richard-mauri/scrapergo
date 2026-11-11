// Command structureprobe inspects a saved HTML snapshot offline. It neither
// retrieves URLs nor changes ScraperGo's production article extraction.
package main

import (
 "encoding/json"
 "flag"
 "fmt"
 "os"

 "clean-article-reader/internal/sourcepreserve"
)

func main(){
 input:=flag.String("input","","existing local HTML snapshot")
 url:=flag.String("url","","source URL to resolve safe relative image and link URLs")
 structureTrace:=flag.Bool("structure-trace",false,"print content-free per-image group structure instead of article text")
 flag.Parse()
 if *input==""||*url==""||flag.NArg()!=0{
  fmt.Fprintln(os.Stderr,"Usage: go run ./cmd/structureprobe -input /tmp/article.html -url https://publisher.example/story")
  os.Exit(2)
 }
 raw,err:=os.ReadFile(*input)
 if err!=nil{fmt.Fprintln(os.Stderr,"Cannot read HTML:",err);os.Exit(1)}
 if *structureTrace {
  result,err:=sourcepreserve.TraceStructure(string(raw),*url)
  if err!=nil{fmt.Fprintln(os.Stderr,"Cannot inspect structure:",err);os.Exit(1)}
  out:=json.NewEncoder(os.Stdout)
  out.SetIndent("","  ")
  if err:=out.Encode(result);err!=nil{fmt.Fprintln(os.Stderr,"Cannot write trace:",err);os.Exit(1)}
  return
 }
 result,err:=sourcepreserve.Inspect(string(raw),*url)
 if err!=nil{fmt.Fprintln(os.Stderr,"Cannot inspect HTML:",err);os.Exit(1)}
 out:=json.NewEncoder(os.Stdout)
 out.SetIndent("","  ")
 if err:=out.Encode(result);err!=nil{fmt.Fprintln(os.Stderr,"Cannot write report:",err);os.Exit(1)}
}
