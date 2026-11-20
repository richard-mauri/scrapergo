// Package ui owns the ScraperGo browser reader presentation and assets.
// Its template, CSS and JavaScript have no article-extraction dependencies.
package ui

import (
 "bytes"
 "embed"
 "strings"

 suiteui "github.com/richard-mauri/homepage/shared/ui"
 "github.com/richard-mauri/homepage/shared/ui/banner"
)

//go:embed templates/index.html static/reader.css static/reader.js
var assets embed.FS

var templateHTML = mustRead("templates/index.html")
var readerCSS = mustRead("static/reader.css")
var readerJS = mustRead("static/reader.js")

func mustRead(path string) string {
 data, err := assets.ReadFile(path)
 if err != nil { panic(err) }
 return string(data)
}

// Template returns the static HTML skeleton, without a compiled version label.
func Template() string {
 page := strings.Replace(templateHTML,"__READER_CSS__",readerCSS+"\n"+string(suiteui.ButtonCSS)+"\n"+string(banner.CSS(suiteui.DesignCSS)),1)
 return strings.Replace(page,"__READER_JS__",readerJS,1)
}

// Render preserves the existing inline asset delivery and Content Security
// Policy while sourcing each presentation concern from its own asset file.
func Render(version, build string) string {
 page := Template()
 var header bytes.Buffer
 err:=banner.Render(&header,banner.Config{
  AppName:"ScraperGo",Version:version,Build:build,
  Description:"Public articles, without the page clutter.",
  HomeURL:"https://home.themauris.org/",TannerURL:"/tanner.jpg",
  AccountState:banner.BrowserOnly,
 })
 if err!=nil {panic(err)} // fixed local configuration, checked by tests
 return strings.Replace(page,"__SUITE_BANNER__",header.String(),1)
}

// RenderWithActivity adds an anonymous, same-origin heartbeat only when the
// server has an explicitly configured private reporter. No service key enters
// this HTML. The existing inline-script CSP continues to apply.
func RenderWithActivity(version,build,browserScript string) string {
 page:=Render(version,build)
 if browserScript=="" {return page}
 return strings.Replace(page,"</body>","<script>"+string(banner.ScriptJS())+"\n"+browserScript+"</script></body>",1)
}
