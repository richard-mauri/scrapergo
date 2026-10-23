// Package ui owns the ScraperGo browser reader presentation and assets.
// Its template, CSS and JavaScript have no article-extraction dependencies.
package ui

import (
 "embed"
 "strings"
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
func Template() string { return templateHTML }

// Render preserves the existing inline asset delivery and Content Security
// Policy while sourcing each presentation concern from its own asset file.
func Render(version, build string) string {
 page := strings.Replace(templateHTML,"__READER_CSS__",readerCSS,1)
 page = strings.Replace(page,"__READER_JS__",readerJS,1)
 return strings.Replace(page,"__APP_VERSION__","v"+version+" · "+build,1)
}
