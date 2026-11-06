// readabilityprobe compares source HTML structure with the HTML and text
// produced by the same Readability version used by ScraperGo.
// It does not make network requests or depend on ScraperGo's browser UI.
package main

import (
 "bytes"
 "encoding/json"
 "flag"
 "fmt"
 "io"
 "net/url"
 "os"
 "path/filepath"
 "strings"

 readability "codeberg.org/readeck/go-readability/v2"
 "golang.org/x/net/html"
)

type element struct {
 Tag string `json:"tag"`
 Text string `json:"text,omitempty"`
 DateTime string `json:"datetime,omitempty"`
}
type snapshot struct {
 Headings []element `json:"headings"`
 Times []element `json:"times"`
 Articles int `json:"articles"`
 Sections int `json:"sections"`
}
type report struct {
 Input string `json:"input"`
 URL string `json:"url"`
 Title string `json:"readabilityTitle"`
 Byline string `json:"readabilityByline,omitempty"`
 Source snapshot `json:"source"`
 ReadabilityHTML snapshot `json:"readabilityHTML"`
 Notes []string `json:"notes"`
}

func main() {
 input := flag.String("input", "", "local HTML fixture or captured page (required)")
 base := flag.String("url", "https://example.org/article", "base URL for relative links")
 out := flag.String("out", "", "output directory for diagnostic files (default: stdout summary)")
 flag.Parse()
 if *input == "" { fmt.Fprintln(os.Stderr, "usage: go run ./cmd/readabilityprobe -input testdata/readability/semantic-live.html [-url https://example.org/story] [-out /tmp/readability-probe]"); os.Exit(2) }
 if err := run(*input, *base, *out, os.Stdout); err != nil { fmt.Fprintln(os.Stderr, "readabilityprobe:", err); os.Exit(1) }
}
func run(input, base, out string, w io.Writer) error {
 raw, err := os.ReadFile(input); if err != nil { return err }
 u, err := url.Parse(base); if err != nil || !u.IsAbs() || (u.Scheme!="http" && u.Scheme!="https") { return fmt.Errorf("base URL must be absolute HTTP(S)") }
 parsed, err := readability.FromReader(bytes.NewReader(raw), u); if err != nil { return fmt.Errorf("Readability parse: %w",err) }
 var htmlBuf, textBuf bytes.Buffer
 if err := parsed.RenderHTML(&htmlBuf); err != nil { return fmt.Errorf("RenderHTML: %w",err) }
 if err := parsed.RenderText(&textBuf); err != nil { return fmt.Errorf("RenderText: %w",err) }
 original, err := inspect(bytes.NewReader(raw)); if err != nil { return fmt.Errorf("source DOM: %w",err) }
 cleaned, err := inspect(bytes.NewReader(htmlBuf.Bytes())); if err != nil { return fmt.Errorf("Readability DOM: %w",err) }
 result:=report{Input:input,URL:base,Title:parsed.Title(),Byline:parsed.Byline(),Source:original,ReadabilityHTML:cleaned,Notes:[]string{
  "Compare source.headings/times with readabilityHTML.headings/times to identify structure lost by Readability.",
  "If Readability HTML preserves structure but text loses it, investigate ScraperGo's RenderText paragraph conversion.",
  "Synthetic fixtures are diagnostic examples, not evidence of live publisher compatibility.",
 }}
 output, err := json.MarshalIndent(result,"","  "); if err != nil { return err }
 if out!="" {
  if err:=os.MkdirAll(out,0755); err!=nil {return err}
  files:=map[string][]byte{"source.html":raw,"readability.html":htmlBuf.Bytes(),"readability.txt":textBuf.Bytes(),"report.json":append(output,'\n')}
  for name, data:=range files {if err:=os.WriteFile(filepath.Join(out,name),data,0644); err!=nil {return err}}
 }
 _,err=fmt.Fprintln(w,string(output))
 return err
}
func inspect(r io.Reader) (snapshot,error) {
 root,err:=html.Parse(r); if err!=nil{return snapshot{},err}
 s:=snapshot{Headings:[]element{},Times:[]element{}}
 var walk func(*html.Node)
 walk=func(n *html.Node){
  if n.Type==html.ElementNode {
   switch n.Data {
   case "article": s.Articles++
   case "section": s.Sections++
   case "h1","h2","h3","h4","h5","h6":
    s.Headings=append(s.Headings,element{Tag:n.Data,Text:normalizedText(n)})
   case "time":
    el:=element{Tag:"time",Text:normalizedText(n)}
    for _,a:=range n.Attr {if a.Key=="datetime"{el.DateTime=a.Val;break}}
    s.Times=append(s.Times,el)
   }
  }
  for c:=n.FirstChild;c!=nil;c=c.NextSibling {walk(c)}
 }
 walk(root)
 return s,nil
}
func normalizedText(n *html.Node) string {
 var pieces []string
 var walk func(*html.Node)
 walk=func(x *html.Node){
  if x.Type==html.TextNode {pieces=append(pieces,x.Data)}
  for c:=x.FirstChild;c!=nil;c=c.NextSibling {walk(c)}
 }
 walk(n)
 return strings.Join(strings.Fields(strings.Join(pieces," "))," ")
}
