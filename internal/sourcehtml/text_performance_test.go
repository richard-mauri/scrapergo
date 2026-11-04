package sourcehtml

import (
 "fmt"
 "strings"
 "testing"
)

func TestTextDeeplyNestedLivePage(t *testing.T) {
 // A large text leaf inside many wrapping containers historically triggered
 // repeated full-size regexp substitutions at every ancestor.
 leaf:=&Node{Tag:"#text",Text:strings.Repeat("Candidate update with election context and a named district. ",12000)}
 current:=leaf
 for depth:=0;depth<100;depth++ {
  parent:=&Node{Tag:"div",Kids:[]*Node{current}}
  current.Parent=parent
  current=parent
 }
 got:=Text(current)
 if want:=Normalize(leaf.Text);got!=want {
  t.Fatalf("deep text corrupted: got %d bytes, want %d",len(got),len(want))
 }
}

func TestTextPreservesHierarchyWhitespaceAndBlockedTags(t *testing.T) {
 n:=&Node{Tag:"article",Kids:[]*Node{
  {Tag:"h2",Kids:[]*Node{{Tag:"#text",Text:"Latest  results"}}},
  {Tag:"p",Kids:[]*Node{
   {Tag:"#text",Text:"One  "},
   {Tag:"strong",Kids:[]*Node{{Tag:"#text",Text:"important"}}},
   {Tag:"#text",Text:" update."},
  }},
  {Tag:"script",Kids:[]*Node{{Tag:"#text",Text:"hidden script data"}}},
 }}
 got:=Text(n)
 if got!="Latest results One important update." {
  t.Fatalf("unexpected visible text %q",got)
 }
 if strings.Contains(got,"hidden script") {t.Fatal("blocked script leaked")}
 if got:=Text(&Node{Tag:"#text",Text:"raw  leaf"});got!="raw  leaf" {t.Fatalf("text leaf normalization changed: %q",got)}
 if Text(nil)!="" {t.Fatal("nil text must be empty")}
}

func BenchmarkTextDeepLiveBlog(b *testing.B) {
 leaf:=&Node{Tag:"#text",Text:strings.Repeat("Live update from a district. ",7000)}
 current:=leaf
 for i:=0;i<70;i++ {
  parent:=&Node{Tag:fmt.Sprintf("div"),Kids:[]*Node{current}}
  current=parent
 }
 b.ResetTimer()
 for i:=0;i<b.N;i++ {_=Text(current)}
}
