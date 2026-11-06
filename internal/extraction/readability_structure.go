package extraction

import (
 "strings"

 "clean-article-reader/internal/article"
 "clean-article-reader/internal/sourcehtml"
)

// PreserveReadabilityStructure uses the markup selected by Readability to
// recover semantic headings and standalone timestamps. It does not use URL
// patterns, publisher names, or relative-time phrase recognition.
//
// This stage is intentionally an overlay: it changes a block only when its
// exact normalized text can be tied uniquely to a heading in selected HTML.
// A time element is separated only when directly preceding that heading in
// the selected heading/time sequence and its exact source text matches the
// preceding block's suffix. All existing article prose and links are retained.
func PreserveReadabilityStructure(blocks []article.Block, cleanedHTML string) []article.Block {
 if len(blocks)==0 || strings.TrimSpace(cleanedHTML)=="" {return blocks}
 root:=sourcehtml.Parse(cleanedHTML)
 nodes:=sourcehtml.Find(root,func(n *sourcehtml.Node)bool {
  return n.Tag=="h2" || n.Tag=="h3" || n.Tag=="h4" || n.Tag=="time"
 })
 if len(nodes)==0 {return blocks}
 headingCounts:=map[string]int{}
 blockCounts:=map[string]int{}
 for _,n:=range nodes {
  if n.Tag!="time" {headingCounts[sourcehtml.Normalize(sourcehtml.Text(n))]++}
 }
 for _,b:=range blocks {if b.Kind=="p" || b.Kind=="h2" || b.Kind=="h3" || b.Kind=="h4" {
  blockCounts[sourcehtml.Normalize(b.Text)]++
 }}
 type evidence struct {kind, timestamp string}
 recoverable:=map[string]evidence{}
 for i,n:=range nodes {
  if n.Tag=="time" {continue}
  text:=sourcehtml.Normalize(sourcehtml.Text(n))
  if text=="" || headingCounts[text]!=1 || blockCounts[text]!=1 {continue}
  ev:=evidence{kind:n.Tag}
  if i>0 && nodes[i-1].Tag=="time" {
   ev.timestamp=sourcehtml.Normalize(sourcehtml.Text(nodes[i-1]))
  }
  recoverable[text]=ev
 }
 if len(recoverable)==0 {return blocks}
 result:=make([]article.Block,0,len(blocks)+len(recoverable))
 for _,b:=range blocks {
  key:=sourcehtml.Normalize(b.Text)
  ev,found:=recoverable[key]
  if !found || (b.Kind!="p" && b.Kind!="h2" && b.Kind!="h3" && b.Kind!="h4") {
   result=append(result,b);continue
  }
  // A preexisting heading cannot justify altering unrelated neighboring text.
  if ev.timestamp!="" && len(result)>0 {
   prev:=&result[len(result)-1]
   if prev.Kind=="p" && sourcehtml.Normalize(prev.Text)==ev.timestamp {
    prev.Kind="time"
   } else if prev.Kind=="p" && splitTrailingSourceTime(prev,ev.timestamp) {
    result=append(result,article.Block{Kind:"time",Text:ev.timestamp})
   }
  }
  b.Kind=ev.kind
  result=append(result,b)
 }
 return result
}

// splitTrailingSourceTime removes only the exact HTML time element text at the
// end of a preceding paragraph, with a whitespace boundary. It never guesses
// what looks like a timestamp, and it retains inline links and formatting.
func splitTrailingSourceTime(b *article.Block, stamp string) bool {
 if b==nil || stamp=="" {return false}
 original:=sourcehtml.Normalize(b.Text)
 if !strings.HasSuffix(original,stamp) || len(original)<=len(stamp) {return false}
 boundary:=original[:len(original)-len(stamp)]
 if len(boundary)==0 || boundary[len(boundary)-1]!=' ' {return false}
 prose:=strings.TrimSpace(boundary)
 if prose=="" {return false}
 if len(b.Parts)>0 {
  parts:=append([]article.TextPart(nil),b.Parts...)
  remaining:=len(stamp)
  for i:=len(parts)-1;i>=0 && remaining>0;i-- {
   text:=parts[i].Text
   if len(text)>=remaining {
    if !strings.HasSuffix(strings.TrimRight(text," \t\n"),stamp[len(stamp)-remaining:]) {return false}
    trim:=len(strings.TrimRight(text," \t\n"))-(remaining)
    if trim<0{return false}
    parts[i].Text=strings.TrimRight(text[:trim]," \t\n")
    remaining=0
   } else {
    // Uncommon timestamps spanning styled runs are left untouched rather
    // than risking link truncation.
    return false
   }
  }
  if remaining!=0{return false}
  b.Parts=parts
 }
 b.Text=prose
 return true
}
