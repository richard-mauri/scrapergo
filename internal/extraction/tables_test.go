package extraction

import (
 "clean-article-reader/internal/sourcehtml"
 "strings"
 "testing"
)

func TestTableBlockPreservesYouthSportsGrid(t *testing.T) {
 raw:=`<article><table><thead><tr><th>2027</th><th>2028</th><th>2029</th><th>2030</th></tr></thead>
 <tbody><tr><td>Baseball</td><td>American Football</td><td>Basketball</td><td>Figure Skating</td></tr>
 <tr><td>Golf</td><td>Swimming</td><td>Tennis</td><td><a href="/skiing">Skiing</a></td></tr></tbody></table></article>`
 root:=sourcehtml.Parse(raw)
 tables:=sourcehtml.Find(root,func(n *sourcehtml.Node)bool{return n.Tag=="table"})
 if len(tables)!=1 {t.Fatalf("tables=%d",len(tables))}
 b,ok:=TableBlock(tables[0],"https://example.org/story")
 if !ok || len(b.Rows)!=3 {t.Fatalf("table extraction failed: %#v",b)}
 if !b.Rows[0].Cells[0].Header || b.Rows[0].Cells[3].Text!="2030" {t.Fatal("lost column headers")}
 if b.Rows[1].Cells[1].Text!="American Football" || b.Rows[1].Cells[3].Text!="Figure Skating" {t.Fatal("lost cell alignment")}
 if len(b.Rows[2].Cells[3].Parts)==0 || b.Rows[2].Cells[3].Parts[0].URL!="https://example.org/skiing" {t.Fatal("lost safe inline link")}
}

func TestTableBlockPreservesSpans(t *testing.T) {
 root:=sourcehtml.Parse(`<table><tr><th colspan="2">Event</th></tr><tr><td rowspan="2">Half Dollars</td><td>Basketball</td></tr><tr><td>Rugby</td></tr></table>`)
 tables:=sourcehtml.Find(root,func(n *sourcehtml.Node)bool{return n.Tag=="table"})
 b,ok:=TableBlock(tables[0],"https://example.org")
 if !ok || b.Rows[0].Cells[0].ColSpan!=2 || b.Rows[1].Cells[0].RowSpan!=2 {t.Fatalf("lost spans: %#v",b)}
}

func TestFallbackExtractBlocksKeepsTable(t *testing.T){
 html:=`<article><p>This is an editorial paragraph with sufficient words to be a valid source.</p><table><tr><th>Year</th><th>Sport</th></tr><tr><td>2027</td><td>Baseball</td></tr></table><p>Another editorial paragraph following the table has good content.</p></article>`
 blocks:=ExtractBlocks(sourcehtml.Parse(html),"https://example.org")
 found:=false
 for _,b:=range blocks {if b.Kind=="table" {found=true;if len(b.Rows)!=2 || !strings.Contains(b.Text,"Baseball"){t.Fatalf("bad table: %#v",b)}}}
 if !found {t.Fatal("table discarded by fallback extractor")}
}
