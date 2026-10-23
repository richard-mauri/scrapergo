package reader_test

import (
 articlemodel "clean-article-reader/internal/article"
 "clean-article-reader/internal/bootstrap"
 "clean-article-reader/internal/recovery"
 "clean-article-reader/internal/reader"
 "clean-article-reader/internal/structured"
 "os"
 "strings"
 "testing"
)

// These fixed, wholly synthetic sources characterize extraction behavior.
// They are NOT archived captures of real publishers.
func TestFrozenStructuralLeadContract(t *testing.T) {
 raw,err:=os.ReadFile("../../testdata/structural_lead_contract.html")
 if err!=nil {t.Fatal(err)}
 cleaned:=structured.StripEmbeddedMedia(string(raw))
 extracted,err:=bootstrap.ExtractWithEvidence(cleaned,"https://news.example.test/report")
 if err!=nil {t.Fatalf("extract fixture: %v",err)}
 if extracted.Article.WordCount<60 {t.Fatalf("unexpectedly short extraction: %d",extracted.Article.WordCount)}
 if extracted.Evidence.MetadataImageURL!="https://media.example.test/preview.jpg" {
  t.Fatalf("source metadata lost: %q",extracted.Evidence.MetadataImageURL)
 }
 assembled:=reader.DefaultProcessor().Process(extracted,cleaned)
 if !strings.Contains(assembled.ImageURL,"editorial-lead.jpg") {t.Fatalf("wrong lead: %q; blocks=%+v",assembled.ImageURL,assembled.Blocks)}
 for _,block:=range assembled.Blocks {
  if block.Kind=="img" && strings.Contains(block.URL,"reporter.jpg") && block.Role=="lead" {t.Fatalf("author portrait selected as lead: %+v",block)}
 }
}

func TestFrozenSourceImageRolesContract(t *testing.T) {
 raw,err:=os.ReadFile("../../testdata/source_role_contract.html")
 if err!=nil {t.Fatal(err)}
 article:=articlemodel.Article{Blocks:[]articlemodel.Block{
  {Kind:"img",URL:"https://media.example.test/portrait.jpg"},
  {Kind:"img",URL:"https://media.example.test/field-photo.jpg"},
  {Kind:"img",URL:"https://media.example.test/linked-portrait.jpg"},
 }}
 recovery.ClassifyImagesFromSource(&article,string(raw),"https://news.example.test/story")
 for i,want:=range []string{"author","editorial","author"} {
  if article.Blocks[i].Role!=want {t.Fatalf("image %d role = %q, want %q",i,article.Blocks[i].Role,want)}
 }
 article.Blocks[1].Role="lead"
 recovery.ClassifyImagesFromSource(&article,string(raw),"https://news.example.test/story")
 if article.Blocks[1].Role!="lead" {t.Fatal("classification replaced selected lead")}
}
