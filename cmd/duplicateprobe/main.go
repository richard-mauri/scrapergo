// Command duplicateprobe inspects an already saved publisher HTML capture.
// It makes no network requests, does not persist article content, and uses
// ScraperGo's current extraction and reader implementations.
package main

import (
 "crypto/sha256"
 "encoding/hex"
 "flag"
 "fmt"
 "os"
 "sort"
 "strings"
 "unicode"

 "clean-article-reader/internal/article"
 "clean-article-reader/internal/bootstrap"
 "clean-article-reader/internal/extraction"
 "clean-article-reader/internal/reader"
 "clean-article-reader/internal/structured"
)

type match struct {First,Second int; FirstKind,SecondKind, Type, Digest string}

func normalize(s string) string {
 s=strings.Map(func(r rune)rune {
  if unicode.IsSpace(r) {return ' '}
  return unicode.ToLower(r)
 },s)
 return strings.Join(strings.Fields(s)," ")
}
func matches(blocks []article.Block) []match {
 var result []match
 for i,one:=range blocks {
  if one.Kind=="table" || one.Kind=="topic-link" {continue}
  a:=normalize(one.Text)
  if len([]rune(a))<24 {continue}
  for j:=i+1;j<len(blocks);j++ {
   two:=blocks[j]
   if two.Kind=="table" || two.Kind=="topic-link" {continue}
   b:=normalize(two.Text)
   if len([]rune(b))<24 {continue}
   typ:=""
   key:=a
   if a==b {typ="exact"} else if len([]rune(a))>=32 && len([]rune(b))>=32 {
    short,long:=a,b
    if len([]rune(short))>len([]rune(long)) {short,long=long,short}
    if len([]rune(short))>=30 && (strings.HasPrefix(long,short+" ") || strings.HasSuffix(long," "+short)) {
     typ="overlap"
     key=short
    }
   }
   if typ=="" {continue}
   h:=sha256.Sum256([]byte(key))
   result=append(result,match{i+1,j+1,one.Kind,two.Kind,typ,hex.EncodeToString(h[:6])})
  }
 }
 return result
}
func main() {
 file:=flag.String("input","","existing private publisher HTML capture")
 url:=flag.String("url","","publisher article URL used only to resolve source links")
 evidence:=flag.Bool("source-evidence",false,"compare match presence in source and Readability selected HTML")
 provenance:=flag.Bool("source-provenance",false,"report per-block source and Readability element ancestry and DOM order")
 flag.Parse()
 if *file=="" || *url=="" || flag.NArg()!=0 {
  fmt.Fprintln(os.Stderr,"Usage: go run ./cmd/duplicateprobe -input /tmp/article.html -url https://publisher.example/article")
  os.Exit(2)
 }
 raw,err:=os.ReadFile(*file)
 if err!=nil {fmt.Fprintln(os.Stderr,"Cannot open HTML capture:",err);os.Exit(1)}
 cleaned:=structured.StripEmbeddedMedia(string(raw))
 trace:=extraction.NewRecoveryTrace()
 result,err:=bootstrap.ExtractWithEvidenceTraced(cleaned,*url,trace)
 if err!=nil {fmt.Fprintln(os.Stderr,"Extraction failed:",err);os.Exit(1)}
 final:=reader.DefaultProcessor().Process(result,cleaned)
 trace.ObserveTransition("reader-postprocessing",result.Article.Blocks,final.Blocks)
 fmt.Printf("Offline result: %d canonical blocks, method %s; %d candidate pairs\n",len(final.Blocks),final.Method,len(matches(final.Blocks)))
 fmt.Println("Stage                         Before/after exact  Before/after overlap")
 for _,s:=range trace.Results() {
  if s.Outcome=="not_reached" {continue}
  fmt.Printf("%-29s %5d / %-5d       %5d / %-5d\n",s.Stage,s.BeforeExactPairs,s.AfterExactPairs,s.BeforeOverlapPairs,s.AfterOverlapPairs)
 }
 pairs:=matches(final.Blocks)
 sort.Slice(pairs,func(i,j int)bool {
  if pairs[i].First!=pairs[j].First{return pairs[i].First<pairs[j].First}
  return pairs[i].Second<pairs[j].Second
 })
 fmt.Println("Final candidate pairs: block indices are 1-based; digest identifies matching text locally")
 for _,p:=range pairs {
  fmt.Printf("  %d(%s) / %d(%s) %-7s text-hash=%s\n",p.First,p.FirstKind,p.Second,p.SecondKind,p.Type,p.Digest)
 }
 if *evidence {
  if err:=reportSourceEvidence(cleaned,*url,final.Blocks,pairs);err!=nil {
   fmt.Fprintln(os.Stderr,"Source evidence:",err);os.Exit(1)
  }
 }
 if *provenance {
  if err:=reportSourceProvenance(cleaned,*url,final.Blocks,pairs);err!=nil {
   fmt.Fprintln(os.Stderr,"Source provenance:",err);os.Exit(1)
  }
 }
 fmt.Println("No source URL, text, credentials or HTML has been written to disk by this command.")
}
