package main

import (
 "fmt"
 "strings"

 "clean-article-reader/internal/article"
 "clean-article-reader/internal/sourcehtml"
)

// captionShape reports which candidate pattern a pair could satisfy, without
// displaying article text or asserting that either block is a caption.
// It never authorizes changes to production extraction.
type captionShape struct {
 Gap int
 CaptionPlusMiddleExact bool
 CaptionPlusMiddleCompact bool
 PairExact bool
 PairCompact bool
 MiddleKind string
}

func classifyCaptionPair(blocks []article.Block, first, second int) captionShape {
 out:=captionShape{Gap:-1,MiddleKind:"none"}
 if first<1 || second>len(blocks) || first>=second {return out}
 a,b:=blocks[first-1],blocks[second-1]
 out.Gap=second-first-1
 compact:=func(s string)string {
  return strings.ToLower(strings.Join(strings.Fields(sourcehtml.Normalize(s)),""))
 }
 x,y:=sourcehtml.Normalize(a.Text),sourcehtml.Normalize(b.Text)
 out.PairExact=x==y
 out.PairCompact=compact(x)==compact(y)
 if out.Gap==1 {
  mid:=blocks[first]
  out.MiddleKind=mid.Kind
  combined:=sourcehtml.Normalize(x+" "+mid.Text)
  out.CaptionPlusMiddleExact=combined==y
  out.CaptionPlusMiddleCompact=compact(combined)==compact(y)
 }
 return out
}

func printCaptionCandidateShapes(blocks []article.Block,pairs []match) {
 fmt.Println("Candidate shape evidence (no source text; NOT a deletion decision)")
 for _,p:=range pairs {
  s:=classifyCaptionPair(blocks,p.First,p.Second)
  fmt.Printf("  pair %d/%d gap=%d middle=%s caption+middle-exact=%t compact=%t pair-exact=%t pair-compact=%t\n",
   p.First,p.Second,s.Gap,s.MiddleKind,
   s.CaptionPlusMiddleExact,s.CaptionPlusMiddleCompact,s.PairExact,s.PairCompact)
 }
 fmt.Println("Shape matches do not satisfy source/selected HTML image ownership or independent editorial-paragraph safeguards.")
}
