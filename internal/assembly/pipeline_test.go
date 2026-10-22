package assembly

import (
 "testing"
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/extraction"
)

func TestRunOrderAndIsolation(t *testing.T) {
 input:=extraction.Result{Article:article.Article{Title:"Original",Blocks:[]article.Block{{Kind:"p",Text:"original"}}},Evidence:extraction.SourceEvidence{SourceURL:"https://example.org/test"}}
 var order []string
 stage:=func(name string,fn func(*article.Article)) Stage {
  return StageFunc{StageName:name,Transform:func(a *article.Article,_ extraction.SourceEvidence,_ string){order=append(order,name);fn(a)}}
 }
 result:=Run(input,"HTML",
  stage("first",func(a *article.Article){a.Blocks[0].Text="updated"}),
  stage("second",func(a *article.Article){a.Blocks=append(a.Blocks,article.Block{Kind:"p",Text:"second"})}),
 )
 if len(order)!=2||order[0]!="first"||order[1]!="second" {t.Fatalf("stage order: %v",order)}
 if input.Article.Blocks[0].Text!="original"||len(input.Article.Blocks)!=1 {t.Fatalf("mutated initial extraction: %+v",input.Article)}
 if result.Blocks[0].Text!="updated"||len(result.Blocks)!=2 {t.Fatalf("bad result: %+v",result)}
}
