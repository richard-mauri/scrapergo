// Package app owns ScraperGo HTTP routing, server lifecycle and listener configuration.
package app

import (
 "context"
 "os"
 "errors"
 "fmt"
 "log"
 "net/http"
 "syscall"
 "time"

 "clean-article-reader/internal/ui"
 "clean-article-reader/internal/activity"
)

type Handlers struct {
 Read http.Handler
 PDF http.Handler
 Classify http.HandlerFunc
 ImageMetadata http.Handler
 ArticleImage http.HandlerFunc
 Thumbnail http.HandlerFunc
 DiagnosticExtract http.HandlerFunc
 DiagnosticImageTrace http.HandlerFunc
 DiagnosticRawHTML http.HandlerFunc
 DiagnosticClassificationTrace http.HandlerFunc
 TannerPhoto []byte
}
type Config struct {
 Port string
 Version string
 Build string
 Activity *activity.ActivityReporter
}

func Router(h Handlers, config Config) http.Handler {
 mux:=http.NewServeMux()
 if config.Activity!=nil {
  mux.HandleFunc("/api/activity/ping",config.Activity.Ping)
 }
 mux.Handle("/api/read",h.Read)
 if h.PDF != nil { mux.Handle("/api/pdf",h.PDF) }
 mux.HandleFunc("/api/classify",h.Classify)
 mux.Handle("/api/image-metadata",h.ImageMetadata)
 mux.HandleFunc("/api/article-image",h.ArticleImage)
 mux.HandleFunc("/api/thumbnail",h.Thumbnail)
 mux.HandleFunc("/api/diagnostic/extract",h.DiagnosticExtract)
 if h.DiagnosticRawHTML != nil { mux.HandleFunc("/api/diagnostic/raw-html",h.DiagnosticRawHTML) }
 mux.HandleFunc("/api/diagnostic/image-trace",h.DiagnosticImageTrace)
 if h.DiagnosticClassificationTrace != nil {mux.HandleFunc("/api/diagnostic/classification-trace",h.DiagnosticClassificationTrace)}
 mux.HandleFunc("/tanner.jpg",func(w http.ResponseWriter,r *http.Request) {
  if r.Method!=http.MethodGet && r.Method!=http.MethodHead {
   w.Header().Set("Allow","GET, HEAD")
   http.Error(w,"Method not allowed",http.StatusMethodNotAllowed)
   return
  }
  w.Header().Set("Content-Type","image/jpeg")
  w.Header().Set("Cache-Control","public, max-age=86400")
  w.Write(h.TannerPhoto)
 })
 mux.HandleFunc("/",func(w http.ResponseWriter,r *http.Request) {
  if r.URL.Path!="/" {http.NotFound(w,r);return}
  w.Header().Set("Content-Type","text/html; charset=utf-8")
  w.Header().Set("Content-Security-Policy","default-src 'self'; img-src 'self' https: data:; style-src 'unsafe-inline'; script-src 'unsafe-inline'; base-uri 'none'; object-src 'none'; frame-src 'self'")
  if config.Activity!=nil {fmt.Fprint(w,ui.RenderWithActivity(config.Version,config.Build,activity.BrowserScript()))} else {fmt.Fprint(w,ui.Render(config.Version,config.Build))}
 })
 if config.Activity==nil {return mux}
 // Count only application work; never count lightweight presence observations,
 // the UI bootstrap, or the static Tanner image as in-flight extraction work.
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.URL.Path=="/" || r.URL.Path=="/tanner.jpg" || r.URL.Path=="/api/activity/ping" {
   mux.ServeHTTP(w,r);return
  }
  done:=config.Activity.Working()
  defer done()
  mux.ServeHTTP(w,r)
 })
}

func Run(config Config, handlers Handlers) error {
 reporter,err:=activity.NewActivityReporter(os.Getenv("SCRAPERGO_ACTIVITY_URL"),os.Getenv("HOMEPAGE_ACTIVITY_KEY_SCRAPERGO"))
 if err!=nil {log.Print("ScraperGo activity reporting disabled: invalid configuration")}
 if err==nil && reporter!=nil {
  config.Activity=reporter
  ctx,cancel:=context.WithCancel(context.Background())
  defer cancel()
  go reporter.Run(ctx)
 }
 server:=&http.Server{Addr:":"+config.Port,Handler:Router(handlers,config),ReadHeaderTimeout:5*time.Second}
 log.Printf("ScraperGo v%s · %s listening on http://localhost:%s",config.Version,config.Build,config.Port)
 if err:=server.ListenAndServe();err!=nil {
  if errors.Is(err,syscall.EADDRINUSE) {
   log.Printf("Port %s is already in use. Try a different port:\n  go run . --port 8081\n  ./dist/scrapergo-darwin-amd64 --port 8081\nOr set PORT=8081 before launching.",config.Port)
  }else{log.Printf("HTTP server failed: %v",err)}
  return err
 }
 return nil
}
