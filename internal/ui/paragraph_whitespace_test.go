package ui

import (
 "strings"
 "testing"
)

func TestOrdinaryArticleParagraphsCollapseWhitespace(t *testing.T){
 page:=Render("test-version","test-build")
 if !strings.Contains(page,".content p{margin:1.1em 0;white-space:normal}") {t.Fatal("ordinary paragraphs should collapse source whitespace")}
}
