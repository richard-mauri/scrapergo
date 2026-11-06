// htmlcapture saves publisher HTML using ScraperGo's existing validated
// retrieval boundary. This command is diagnostic only and has no UI or server.
package main

import (
 "context"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "flag"
 "fmt"
 "io"
 "os"
 "path/filepath"
 "strings"
 "time"

 "clean-article-reader/internal/fetcher"
)

type captureMetadata struct {
 RequestedURL string `json:"requestedURL"`
 FinalURL string `json:"finalURL"`
 RetrievedAt string `json:"retrievedAt"`
 Bytes int `json:"bytes"`
 SHA256 string `json:"sha256"`
 Retrieval string `json:"retrieval"`
}

type downloader func(context.Context, string) (string, string, error)

func main() {
 address := flag.String("url", "", "HTTP(S) page URL to capture (required)")
 output := flag.String("out", "", "new local .html output file (required; never overwrites)")
 flag.Parse()
 if *address == "" || *output == "" || flag.NArg() != 0 {
  fmt.Fprintln(os.Stderr,"usage: go run ./cmd/htmlcapture -url https://example.org/article -out /tmp/article.html")
  os.Exit(2)
 }
 ctx, cancel := context.WithTimeout(context.Background(),45*time.Second)
 defer cancel()
 result,err := capture(ctx,*address,*output,fetcher.Download,time.Now)
 if err!=nil {fmt.Fprintln(os.Stderr,"htmlcapture:",err);os.Exit(1)}
 fmt.Printf("Saved %d bytes to %s; metadata: %s.json; SHA-256: %s\n",result.Bytes,*output,result.SHA256)
}

// capture always delegates network access to internal/fetcher.Download. No
// custom HTTP transport, redirect exceptions, cookies, or access bypass exists.
func capture(ctx context.Context, address, output string, download downloader, now func() time.Time) (captureMetadata,error) {
 if strings.TrimSpace(address)=="" || strings.TrimSpace(output)=="" {return captureMetadata{},errors.New("URL and output path are required")}
 if filepath.Ext(output)!=".html" && filepath.Ext(output)!=".htm" {return captureMetadata{},errors.New("output must be an .html or .htm file")}
 // Reject existing destination/metadata (including symlinks) without replacing either.
 for _,name:=range []string{output,output+".json"} {
  _,err:=os.Lstat(name)
  if err==nil {return captureMetadata{},fmt.Errorf("refusing to overwrite %q",name)}
  if !errors.Is(err,os.ErrNotExist){return captureMetadata{},err}
 }
 html, finalURL,err := download(ctx,address)
 if err!=nil{return captureMetadata{},err}
 if len(html)==0{return captureMetadata{},errors.New("publisher returned empty HTML")}
 if len(html)>fetcher.MaxHTML{return captureMetadata{},errors.New("HTML exceeds configured download limit")}
 digest:=sha256.Sum256([]byte(html))
 meta:=captureMetadata{
  RequestedURL: address, FinalURL:finalURL,
  RetrievedAt:now().UTC().Format(time.RFC3339Nano),
  Bytes:len(html),SHA256:hex.EncodeToString(digest[:]),
  Retrieval:"ScraperGo validated fetcher.Download (success; final HTTP status not exposed)",
 }
 metadata,err:=json.MarshalIndent(meta,"","  ")
 if err!=nil{return captureMetadata{},err}
 // O_EXCL ensures a concurrent file creation cannot be overwritten. Local
 // captures are private by default and never printed to the console.
 file,err:=os.OpenFile(output,os.O_WRONLY|os.O_CREATE|os.O_EXCL,0600)
 if err!=nil{return captureMetadata{},err}
 saved:=false
 defer func(){if !saved {_=os.Remove(output)}}()
 if _,err=io.WriteString(file,html);err!=nil {_=file.Close();return captureMetadata{},err}
 if err=file.Close();err!=nil{return captureMetadata{},err}
 sidecar,err:=os.OpenFile(output+".json",os.O_WRONLY|os.O_CREATE|os.O_EXCL,0600)
 if err!=nil{return captureMetadata{},err}
 if _,err=sidecar.Write(append(metadata,'\n'));err!=nil {_=sidecar.Close();_ = os.Remove(output+".json");return captureMetadata{},err}
 if err=sidecar.Close();err!=nil {_=os.Remove(output+".json");return captureMetadata{},err}
 saved=true
 return meta,nil
}
