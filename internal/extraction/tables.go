package extraction

import (
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/sourcehtml"
 "strconv"
 "strings"
)

// TableBlock retains the source row/column relationships without serializing HTML.
func TableBlock(n *sourcehtml.Node, baseURL string) (article.Block, bool) {
 b := article.Block{Kind:"table"}
 var rows []article.TableRow
 for _, tr := range sourcehtml.Find(n, func(x *sourcehtml.Node) bool { return x.Tag=="tr" }) {
  // Ignore rows belonging to nested tables.
  nested := false
  for p:=tr.Parent; p!=nil && p!=n; p=p.Parent { if p.Tag=="table" { nested=true; break } }
  if nested { continue }
  row:=article.TableRow{}
  for _, cell := range tr.Kids {
   if cell.Tag!="td" && cell.Tag!="th" { continue }
   c:=article.TableCell{Text:sourcehtml.Normalize(sourcehtml.Text(cell)),Header:cell.Tag=="th",Parts:InlineTextParts(cell,baseURL)}
   if v,e:=strconv.Atoi(cell.Attr["colspan"]); e==nil && v>1 && v<=100 { c.ColSpan=v }
   if v,e:=strconv.Atoi(cell.Attr["rowspan"]); e==nil && v>1 && v<=100 { c.RowSpan=v }
   row.Cells=append(row.Cells,c)
  }
  if len(row.Cells)>0 {rows=append(rows,row)}
 }
 if len(rows)<2 {return b,false}
 b.Rows=rows
 var lines []string
 for _,row:=range rows {var cells []string;for _,c:=range row.Cells {cells=append(cells,c.Text)};lines=append(lines,strings.Join(cells," | "))}
 b.Text=strings.Join(lines,"\n")
 return b,true
}

// RestoreSourceTables replaces recognizable flattened table cells in the
// Readability text result. A table must have multiple matching cell values;
// incidental navigation tables are not injected into unrelated article text.
func RestoreSourceTables(blocks []article.Block, raw, source string) []article.Block {
 root:=sourcehtml.Parse(raw)
 tables:=sourcehtml.Find(root,func(n *sourcehtml.Node) bool{return n.Tag=="table"})
 for _,table:=range tables {
  if sourcehtml.Noisy(table) {continue}
  tb,ok:=TableBlock(table,source);if !ok {continue}
  values:=map[string]bool{}
  for _,r:=range tb.Rows {for _,c:=range r.Cells {if len(strings.TrimSpace(c.Text))>=2 {values[sourcehtml.Normalize(c.Text)]=true}}}
  matches:=0
  first:=-1
  for i,b:=range blocks {if b.Kind=="table" {continue}; if values[sourcehtml.Normalize(b.Text)] {matches++;if first<0 {first=i}}}
  if first<0 || matches<3 {continue}
  // If Readability flattened a table as one multiline paragraph, preserving
  // the table still improves readability, but avoid deleting surrounding prose.
  output:=make([]article.Block,0,len(blocks)+1)
  for i,b:=range blocks {
   if i==first {output=append(output,tb)}
   if values[sourcehtml.Normalize(b.Text)] && (b.Kind=="p"||b.Kind=="h2"||b.Kind=="h3"||b.Kind=="h4") {continue}
   output=append(output,b)
  }
  blocks=output
 }
 return blocks
}
