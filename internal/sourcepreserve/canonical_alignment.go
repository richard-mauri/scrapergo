package sourcepreserve

import "clean-article-reader/internal/article"

// CanonicalImageAlignment compares actual final reader image blocks with the
// Readability-selected semantic tree. The standalone Article.ImageURL is
// reported separately, since it may intentionally duplicate or replace a
// block image. Exact resolved URL comparison does not establish visual identity.
type CanonicalImageAlignment struct {
 CanonicalImageBlocks int `json:"canonicalImageBlocks"`
 SelectedImages int `json:"selectedImages"`
 ExactURLMatches int `json:"exactURLMatches"`
 CanonicalOnly int `json:"canonicalOnly"`
 SelectedOnly int `json:"selectedOnly"`
 OrderPreserved bool `json:"orderPreserved"`
 CanonicalLeadImagePresent bool `json:"canonicalLeadImagePresent"`
 Matches []ImageMatch `json:"matches,omitempty"`
}

func CompareCanonicalImages(a article.Article,selected SemanticDocument) CanonicalImageAlignment {
 result:=CanonicalImageAlignment{OrderPreserved:true,CanonicalLeadImagePresent:a.ImageURL!=""}
 var source []string
 for _,b:=range a.Blocks {if b.Kind=="img"&&b.URL!=""{source=append(source,b.URL)}}
 result.CanonicalImageBlocks=len(source)
 target:=selectedImages(selected.Nodes)
 result.SelectedImages=len(target)
 positions:=map[string][]int{}
 for i,url:=range source {positions[url]=append(positions[url],i+1)}
 last:=0
 for _,image:=range target{
  queue:=positions[image.SourceURL]
  if len(queue)==0 {result.SelectedOnly++;continue}
  index:=queue[0]
  positions[image.SourceURL]=queue[1:]
  result.Matches=append(result.Matches,ImageMatch{SourceIndex:index,SelectedIndex:image.Ordinal})
  result.ExactURLMatches++
  if index<last {result.OrderPreserved=false}
  last=index
 }
 result.CanonicalOnly=len(source)-result.ExactURLMatches
 return result
}
