package main

import (
 "clean-article-reader/internal/app"
 "clean-article-reader/internal/bootstrap"
 _ "embed"
 "flag"
 "fmt"
 "os"
)

const appVersion = "0.2.85"
var appBuild = "Build 000000"

func appDisplayVersion() string { return "ScraperGo v" + appVersion + " · " + appBuild }

//go:embed tanner.jpg
var tannerJPG []byte

func main() {
 if len(os.Args)==2 && (os.Args[1]=="--version" || os.Args[1]=="-version") {
  fmt.Println(appDisplayVersion())
  return
 }
 port,err:=app.ResolvePort(os.Args[1:],os.Getenv("PORT"),os.Stderr)
 if err==flag.ErrHelp { return }
 if err!=nil {
  fmt.Fprintln(os.Stderr,"ScraperGo:",err)
  os.Exit(2)
 }
 if err=bootstrap.Start(port,appVersion,appBuild,tannerJPG);err!=nil { os.Exit(1) }
}
