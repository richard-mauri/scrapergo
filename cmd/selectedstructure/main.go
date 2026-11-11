// Command selectedstructure compares Readability's selected HTML semantics with
// its flattened text on the same locally saved source. It never fetches URLs.
package main

import (
 "bytes"
 "encoding/json"
 "flag"
 "fmt"
 "os"
 "net/url"
 "strings"

 "clean-article-reader/internal/bootstrap"
 "clean-article-reader/internal/extraction"
 "clean-article-reader/internal/reader"
 "clean-article-reader/internal/structured"

 readability "codeberg.org/readeck/go-readability/v2"
 "clean-article-reader/internal/sourcepreserve"
)

type summary struct {
 Status string `json:"status"`
 SelectedHTMLKinds map[string]int `json:"selectedHTMLKinds"`
 SelectedHTMLImages int `json:"selectedHTMLImages"`
 UnsafeSelectedURLs int `json:"unsafeSelectedURLs,omitempty"`
 FlatTextCharacters int `json:"flatTextCharacters"`
 ReadabilityTitlePresent bool `json:"readabilityTitlePresent"`
 SourceSelectedImageAlignment sourcepreserve.ImageAlignment `json:"sourceSelectedImageAlignment"`
 BaselineImageAlignment sourcepreserve.CanonicalImageAlignment `json:"baselineImageAlignment"`
 BaselineMethod string `json:"baselineMethod,omitempty"`
 BaselineImageStages []imageStage `json:"baselineImageStages,omitempty"`
}
// imageStage uses the production extraction trace's already-established
// stage counters; no publisher text or image URL is retained.
type imageStage struct {
 Stage string `json:"stage"`
 Outcome string `json:"outcome"`
 BeforeImages int `json:"beforeImages"`
 AfterImages int `json:"afterImages"`
}
func imageStages(results []extraction.RecoveryStageResult) []imageStage {
 var stages []imageStage
 for _,r:=range results {
  if r.Outcome=="not_reached" || r.Outcome=="" {continue}
  stages=append(stages,imageStage{Stage:r.Stage,Outcome:r.Outcome,BeforeImages:r.BeforeImages,AfterImages:r.AfterImages})
 }
 return stages
}
func countKinds(nodes []sourcepreserve.SemanticNode,m map[string]int) {
 for _,n:=range nodes {m[n.Kind]++;countKinds(n.Children,m)}
}
func main(){
 input:=flag.String("input","","existing local HTML file (required)")
 base:=flag.String("url","","article base URL for safe relative links (required)")
 flag.Parse()
 if *input==""||*base==""||flag.NArg()!=0 {fmt.Fprintln(os.Stderr,"usage: go run ./cmd/selectedstructure -input /tmp/article.html -url https://example.org/article");os.Exit(2)}
 raw,err:=os.ReadFile(*input);if err!=nil {fmt.Fprintln(os.Stderr,err);os.Exit(1)}
 parsedURL,err:=url.Parse(*base)
 if err!=nil || (parsedURL.Scheme!="http"&&parsedURL.Scheme!="https") || parsedURL.Hostname()=="" || parsedURL.User!=nil {
  fmt.Fprintln(os.Stderr,"invalid HTTP(S) base URL");os.Exit(2)
 }
 // Use the same preprocessing that production uses for this snapshot.
 // All three comparison paths operate on the same local input.
 cleaned:=structured.StripEmbeddedMedia(string(raw))
 parsed,err:=readability.FromReader(strings.NewReader(cleaned),parsedURL)
 if err!=nil {fmt.Fprintln(os.Stderr,"Readability parse:",err);os.Exit(1)}
 var chosenHTML,flatText bytes.Buffer
 if err:=parsed.RenderHTML(&chosenHTML);err!=nil {fmt.Fprintln(os.Stderr,"RenderHTML:",err);os.Exit(1)}
 if err:=parsed.RenderText(&flatText);err!=nil {fmt.Fprintln(os.Stderr,"RenderText:",err);os.Exit(1)}
 tree,err:=sourcepreserve.ConvertSelectedHTML(chosenHTML.String(),*base)
 if err!=nil {fmt.Fprintln(os.Stderr,"selected DOM:",err);os.Exit(1)}
 report:=summary{Status:"offline-comparison",SelectedHTMLKinds:map[string]int{},SelectedHTMLImages:tree.ObservedImages,UnsafeSelectedURLs:tree.UnsafeURLs,FlatTextCharacters:len([]rune(flatText.String())),ReadabilityTitlePresent:parsed.Title()!=""}
 countKinds(tree.Nodes,report.SelectedHTMLKinds)
 report.SourceSelectedImageAlignment,err=sourcepreserve.CompareImageAlignment(cleaned,*base,tree)
 if err!=nil {fmt.Fprintln(os.Stderr,"source/selected image match:",err);os.Exit(1)}
 trace:=extraction.NewRecoveryTrace()
 extracted,err:=bootstrap.ExtractWithEvidenceTraced(cleaned,*base,trace)
 if err!=nil {fmt.Fprintln(os.Stderr,"production baseline extraction:",err);os.Exit(1)}
 canonical:=reader.DefaultProcessor().Process(extracted,cleaned)
 trace.ObserveTransition("reader-postprocessing",extracted.Article.Blocks,canonical.Blocks)
 report.BaselineImageStages=imageStages(trace.Results())
 report.BaselineImageAlignment=sourcepreserve.CompareCanonicalImages(canonical,tree)
 report.BaselineMethod=canonical.Method
 enc:=json.NewEncoder(os.Stdout);enc.SetIndent("","  ")
 if err:=enc.Encode(report);err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(1)}
}
