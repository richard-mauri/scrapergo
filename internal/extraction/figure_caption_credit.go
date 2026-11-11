package extraction

import (
 "strings"

 "clean-article-reader/internal/article"
 "clean-article-reader/internal/sourcehtml"
)

// reconcileImageCaptionCreditTriples removes one redundant *combined* text
// block only when the original source AND Readability-selected HTML both
// attach that same caption/credit text to an actual image.
//
// RenderText may flatten one photo's caption and credit as three consecutive
// blocks: caption, credit, and caption+credit. The two distinct pieces remain
// as separate article blocks; only the redundant combined representation is
// removed. This is not a general paragraph deduplicator.
//
// Source proof must come from a compact single-photo subtree, not a whole
// article, a two-photo gallery, or unrelated text somewhere on the page.
// Ambiguous source markup is deliberately left unchanged.
func reconcileImageCaptionCreditTriples(blocks []article.Block, sourceHTML, selectedHTML string) []article.Block {
 if len(blocks)<3 || sourceHTML=="" || selectedHTML=="" {return blocks}
 // Most articles do not have the exact caption / credit / combined pattern.
 // Avoid parsing source HTML twice for such articles, including long feeds.
 candidates:=[]int{}
 for i:=2;i<len(blocks);i++ {
  if isImageCaptionCreditTriple(blocks[i-2],blocks[i-1],blocks[i]) {
   candidates=append(candidates,i)
  }
 }
 if len(candidates)==0 {return blocks}
 source:=sourcehtml.Parse(sourceHTML)
 selected:=sourcehtml.Parse(selectedHTML)
 if source==nil || selected==nil {return blocks}
 skip:=map[int]bool{}
 for _,i:=range candidates {
  if hasImageCaptionCreditScope(source,blocks[i-2].Text,blocks[i-1].Text,false) &&
   hasImageCaptionCreditScope(selected,blocks[i-2].Text,blocks[i-1].Text,true) {
   // Preserve the independent caption and credit. Images remain in the
   // existing figure positioning and postprocessing pipeline.
   skip[i]=true
  }
 }
 if len(skip)==0 {return blocks}
 result:=make([]article.Block,0,len(blocks)-len(skip))
 for i,b:=range blocks {
  if !skip[i] {result=append(result,b)}
 }
 return result
}

func isImageCaptionCreditTriple(caption,credit,combined article.Block) bool {
 if caption.Kind!="p" || credit.Kind!="p" || combined.Kind!="p" {return false}
 // Do not discard rich text, links or metadata that the concatenated block
 // might uniquely retain.
 if len(combined.Parts)>0 || len(combined.Links)>0 {return false}
 a:=sourcehtml.Normalize(caption.Text)
 b:=sourcehtml.Normalize(credit.Text)
 c:=sourcehtml.Normalize(combined.Text)
 if len([]rune(a))<24 || len([]rune(b))<5 || a==b {return false}
 return c==a+" "+b
}

// hasImageCaptionCreditScope verifies both pieces co-occur in a compact
// single-image branch and are not just coincidental text in an entire article.
// The image must be a direct child or grandchild of the selected container,
// and a single non-image sibling must hold the caption and credit. This is
// structural evidence from an editorial image group, not URL/site heuristics.
func hasImageCaptionCreditScope(root *sourcehtml.Node,caption,credit string,allowParagraphAlternative bool) bool {
 target:=sourcehtml.Normalize(caption+" "+credit)
 if target=="" {return false}
 owners:=map[*sourcehtml.Node]bool{}
 for _,n:=range sourcehtml.Find(root,func(n *sourcehtml.Node)bool {
  return n.Tag=="div" || n.Tag=="figure" || n.Tag=="section"
 }) {
  if sourcehtml.BlockedTags[n.Tag] {continue}
  // Never absorb a broad article paragraph or multi-image gallery.
  imageNodes:=sourcehtml.Find(n,func(x *sourcehtml.Node)bool{return x.Tag=="img"})
  if len(imageNodes)!=1 {continue}
  // Require image/caption siblings rather than content anywhere in a page.
  imageBranch:=false
  textBranch:=false
  for _,child:=range n.Kids {
   if child.Tag=="#text" || sourcehtml.BlockedTags[child.Tag] {continue}
   if len(sourcehtml.Find(child,func(x *sourcehtml.Node)bool{return x.Tag=="img"}))>0 {
    imageBranch=true
    continue
   }
   if !allowParagraphAlternative && len(sourcehtml.Find(child,func(x *sourcehtml.Node)bool{return x.Tag=="p"}))>0 {continue}
   if containsCompleteCaptionCredit(child,target) {textBranch=true}
  }
  // The image branch must contain the same caption+credit, not merely
  // be an unrelated photograph alongside a matching prose paragraph.
  imageCaption:=false
  for _,child:=range n.Kids {
   if child.Tag=="#text" || sourcehtml.BlockedTags[child.Tag] {continue}
   if len(sourcehtml.Find(child,func(x *sourcehtml.Node)bool{return x.Tag=="img"}))==1 &&
    containsCompleteCaptionCredit(child,target) {
    imageCaption=true
   }
  }
  if imageBranch && imageCaption && textBranch {owners[imageNodes[0]]=true}
 }
 return len(owners)==1
}

func containsCompleteCaptionCredit(n *sourcehtml.Node,target string) bool {
 if n==nil || sourcehtml.BlockedTags[n.Tag] {return false}
 // Only a fairly compact caption wrapper is eligible. A nested article
 // containing editorial paragraph or heading elements is not a caption.
 if n.Tag=="article" || n.Tag=="main" || n.Tag=="section" {return false}
 paragraphs:=0
 for _,x:=range sourcehtml.Find(n,func(x *sourcehtml.Node)bool {
  return x.Tag=="p" || x.Tag=="blockquote" || x.Tag=="h1" || x.Tag=="h2" || x.Tag=="h3" || x.Tag=="h4"
 }) {
  if x.Tag=="blockquote" || strings.HasPrefix(x.Tag,"h") {return false}
  paragraphs++
  // A caption can be a single paragraph, but a subtree of multiple
  // editorial paragraphs must not be classified as one media caption.
  if paragraphs>1 {return false}
 }
 text:=sourcehtml.Normalize(sourcehtml.Text(n))
 return text==target || strings.EqualFold(strings.Join(strings.Fields(text),""),strings.Join(strings.Fields(target),""))
}
