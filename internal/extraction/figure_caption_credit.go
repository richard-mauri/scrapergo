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
 if len(blocks)<2 || sourceHTML=="" || selectedHTML=="" {return blocks}
 triples:=[]int{}
 pairs:=[]int{}
 for i:=1;i<len(blocks);i++ {
  if isImageCaptionDuplicatePair(blocks[i-1],blocks[i]) {pairs=append(pairs,i)}
  if i>=2 && isImageCaptionCreditTriple(blocks[i-2],blocks[i-1],blocks[i]) {
   triples=append(triples,i)
  }
 }
 if len(triples)==0 && len(pairs)==0 {return blocks}
 source:=sourcehtml.Parse(sourceHTML)
 selected:=sourcehtml.Parse(selectedHTML)
 if source==nil || selected==nil {return blocks}
 skip:=map[int]bool{}
 for _,i:=range triples {
  if hasImageCaptionCreditScope(source,blocks[i-2].Text,blocks[i-1].Text,false) &&
   (hasImageCaptionCreditScope(selected,blocks[i-2].Text,blocks[i-1].Text,true) ||
    hasFlattenedSelectedCaptionScope(selected,blocks[i-2].Text,blocks[i-1].Text)) {
   skip[i]=true
  }
 }
 for _,i:=range pairs {
  // Both a source image branch and a separate representation must have
  // the same words, independently corroborated in selected Readability HTML.
  // Editorially intentional repeats outside the photo container survive.
  a,b:=blocks[i-1],blocks[i]
  if !hasImageCaptionCreditScope(source,a.Text,"",false) ||
   !hasImageCaptionCreditScope(selected,a.Text,"",true) {continue}
  remove:=i
  // Prefer the version with intact whitespace (notably ". Photographer"
  // rather than ".Photographer"). Equal-quality copies keep first-in-order.
  if len(strings.Fields(b.Text))>len(strings.Fields(a.Text)) {remove=i-1}
  if len(blocks[remove].Parts)>0 || len(blocks[remove].Links)>0 {continue}
  skip[remove]=true
 }
 if len(skip)==0 {return blocks}
 result:=make([]article.Block,0,len(blocks)-len(skip))
 for i,b:=range blocks {if !skip[i] {result=append(result,b)}}
 return result
}

func isImageCaptionDuplicatePair(a,b article.Block) bool {
 if a.Kind!="p" || b.Kind!="p" {return false}
 if len([]rune(a.Text))<30 || len([]rune(b.Text))<30 {return false}
 // A paragraph may differ only because text wrapping dropped whitespace
 // between caption and photographer credit. Exact repetition also qualifies.
 compact:=func(s string)string {return strings.ToLower(strings.Join(strings.Fields(sourcehtml.Normalize(s)),""))}
 return compact(a.Text)==compact(b.Text)
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

// CaptionCreditGateEvidence exposes the exact predicates used by the current
// caption-credit correction for offline diagnostics. It returns only booleans;
// it never returns publisher text, HTML, attributes, image URLs or IDs.
// TripleIndex is 0-based and denotes the final combined block.
type CaptionCreditGateEvidence struct {
 TripleShape bool
 CombinedHasRichContent bool
 SourceImageScope bool
 SelectedImageScope bool
 Eligible bool
}

func InspectCaptionCreditTriple(blocks []article.Block, tripleIndex int, sourceHTML, selectedHTML string) CaptionCreditGateEvidence {
 var result CaptionCreditGateEvidence
 if tripleIndex<2 || tripleIndex>=len(blocks) {return result}
 a,b,c:=blocks[tripleIndex-2],blocks[tripleIndex-1],blocks[tripleIndex]
 result.CombinedHasRichContent=len(c.Parts)>0 || len(c.Links)>0
 result.TripleShape=isImageCaptionCreditTriple(a,b,c)
 if !result.TripleShape || sourceHTML=="" || selectedHTML=="" {return result}
 result.SourceImageScope=hasImageCaptionCreditScope(sourcehtml.Parse(sourceHTML),a.Text,b.Text,false)
 result.SelectedImageScope=hasImageCaptionCreditScope(sourcehtml.Parse(selectedHTML),a.Text,b.Text,true) ||
  hasFlattenedSelectedCaptionScope(sourcehtml.Parse(selectedHTML),a.Text,b.Text)
 result.Eligible=result.SourceImageScope && result.SelectedImageScope
 return result
}

// hasFlattenedSelectedCaptionScope handles Readability's retained paired-photo
// layout when the original caption/credit text appears once inside an image
// branch and once inside a neighboring text branch of the same compact group.
// The publisher source must independently pass hasImageCaptionCreditScope;
// this predicate alone must never authorize removing editorial prose.
func hasFlattenedSelectedCaptionScope(root *sourcehtml.Node,caption,credit string) bool {
 if root==nil {return false}
 target:=sourcehtml.Normalize(caption+" "+credit)
 compact:=func(s string)string{return strings.ToLower(strings.Join(strings.Fields(sourcehtml.Normalize(s)),""))}
 match:=func(s string)bool{return compact(s)==compact(target)}
 allImages:=sourcehtml.Find(root,func(n *sourcehtml.Node)bool{return n.Tag=="img"})
 allParagraphs:=sourcehtml.Find(root,func(n *sourcehtml.Node)bool{return n.Tag=="p"})
 matchedImageOwners:=map[*sourcehtml.Node]bool{}
 for _,img:=range allImages {
  for owner:=img.Parent;owner!=nil && owner.Tag!="article" && owner.Tag!="main" && owner.Tag!="body";owner=owner.Parent {
   // The image must have a unique local caption-bearing wrapper, not
   // a broad gallery's concatenated text.
   if len(sourcehtml.Find(owner,func(n *sourcehtml.Node)bool{return n.Tag=="img"}))!=1 {break}
   if !match(sourcehtml.Text(owner)) {continue}
   // A separate paragraph bearing the complete caption must live outside
   // the image's own wrapper, but inside a shared bounded media group.
   for _,p:=range allParagraphs {
    if !match(sourcehtml.Text(p)) || isNodeWithin(p,owner) {continue}
    for group:=owner.Parent;group!=nil && group.Tag!="article" && group.Tag!="main" && group.Tag!="body";group=group.Parent {
     if !isNodeWithin(p,group) {continue}
     n:=len(sourcehtml.Find(group,func(x *sourcehtml.Node)bool{return x.Tag=="img"}))
     if n>=1 && n<=12 {matchedImageOwners[img]=true}
     break
    }
   }
   break
  }
 }
 return len(matchedImageOwners)==1
}

func isNodeWithin(n,ancestor *sourcehtml.Node) bool {
 for x:=n;x!=nil;x=x.Parent {if x==ancestor{return true}}
 return false
}
