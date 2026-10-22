package links

import (
 "testing"
 "clean-article-reader/internal/article"
)

func TestAssignSelectedPartsPreservesOneToOneMatching(t *testing.T) {
 blocks:=[]article.Block{
  {Kind:"p",Text:"The same paragraph.",Links:[]article.InlineLink{{Text:"old",URL:"https://old.example"}}},
  {Kind:"p",Text:"The same paragraph."},
  {Kind:"img",Text:"The same paragraph."},
  {Kind:"p",Text:"Different paragraph."},
 }
 evidence:=[]SelectedParts{
  {Text:"The same paragraph.",Parts:[]article.TextPart{{Text:"The",URL:"https://one.example"}}},
  {Text:"The same paragraph.",Parts:[]article.TextPart{{Text:"same",URL:"https://two.example"}}},
 }
 AssignSelectedParts(blocks,evidence)
 if len(blocks[0].Parts)!=1||blocks[0].Parts[0].URL!="https://one.example"||len(blocks[0].Links)!=0 {t.Fatalf("first block: %+v",blocks[0])}
 if len(blocks[1].Parts)!=1||blocks[1].Parts[0].URL!="https://two.example" {t.Fatalf("second block: %+v",blocks[1])}
 if len(blocks[2].Parts)!=0||len(blocks[3].Parts)!=0 {t.Fatalf("unselected blocks modified: %+v",blocks)}
}
