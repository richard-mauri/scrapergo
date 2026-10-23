package reader

import (
 "go/parser"
 "go/token"
 "os"
 "path/filepath"
 "strconv"
 "strings"
 "testing"
)

// Guard the first enforced architectural boundary. Remaining package boundaries
// are tightened as their owners migrate in PRs #43-#47.
func TestReaderDoesNotImportPresentationOrTransport(t *testing.T) {
 entries,err:=os.ReadDir(".")
 if err!=nil { t.Fatal(err) }
 for _,entry:=range entries {
  if entry.IsDir() || !strings.HasSuffix(entry.Name(),".go") || strings.HasSuffix(entry.Name(),"_test.go") {continue}
  filename:=filepath.Join(".",entry.Name())
  f,err:=parser.ParseFile(token.NewFileSet(),filename,nil,parser.ImportsOnly)
  if err!=nil {t.Fatal(err)}
  for _,imp:=range f.Imports {
   path,err:=strconv.Unquote(imp.Path.Value)
   if err!=nil {t.Fatal(err)}
   if path=="net/http" || strings.Contains(path,"/internal/httpapi") || strings.Contains(path,"/internal/ui") || strings.Contains(path,"/internal/app") || strings.Contains(path,"/internal/fetcher") {
    t.Errorf("%s: reader imports forbidden presentation, transport, lifecycle or retrieval package %q",filename,path)
   }
  }
 }
}
