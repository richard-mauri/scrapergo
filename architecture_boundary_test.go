package main

import (
 "go/parser"
 "go/ast"
 "go/token"
 "os"
 "path/filepath"
 "strings"
 "testing"
)

// TestProductionRootBoundary prevents HTTP, extraction and diagnostics code
// from drifting back into the executable's bootstrap-only package.
func TestProductionRootBoundary(t *testing.T) {
 entries,err:=os.ReadDir(".")
 if err!=nil {t.Fatal(err)}
 count:=0
 for _,entry:=range entries {
  name:=entry.Name()
  if entry.IsDir()||!strings.HasSuffix(name,".go")||strings.HasSuffix(name,"_test.go"){continue}
  count++
  if name!="main.go"{t.Errorf("production Go code at root: %s (move to internal/)",name)}
 }
 if count!=1 {t.Errorf("expected exactly one production root Go file, got %d",count)}
 file,err:=parser.ParseFile(token.NewFileSet(),"main.go",nil,0)
 if err!=nil {t.Fatal(err)}
 allowed:=map[string]bool{"main":true,"appDisplayVersion":true}
 for _,decl:=range file.Decls{
  if fn,ok:=decl.(*ast.FuncDecl);ok&&!allowed[fn.Name.Name]{
   t.Errorf("unexpected root production function %q",fn.Name.Name)
  }
 }
 for _,imp:=range file.Imports {
  p:=strings.Trim(imp.Path.Value,"\"")
  if strings.HasPrefix(p,"clean-article-reader/internal/")&&p!="clean-article-reader/internal/app"&&p!="clean-article-reader/internal/bootstrap"{
   t.Errorf("main.go should only depend on app and bootstrap: %s",p)
  }
 }
 for _,name:=range []string{"app_wiring.go","extraction_handoff.go","slate_diagnostic.go","structured_document.go"} {
  if _,err:=os.Stat(filepath.Clean(name));err==nil{t.Errorf("obsolete root production adapter %s",name)}
 }
}
