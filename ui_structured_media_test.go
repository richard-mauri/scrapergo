package main

import (
 "clean-article-reader/internal/ui"
 "strings"
 "testing"
)

func TestSingleMediaRenderPath(t *testing.T) {
 page:=ui.Render(appVersion,appBuild)
 if strings.Contains(page,"if(data.imageUrl){const pic="){t.Fatal("independent hero rendering path must be removed")}
 if !strings.Contains(page,"if(b.role==='lead')fig.className='lead-photo'"){t.Fatal("typed lead role not used by renderer")}
 if !strings.Contains(page,"fig.appendChild(im)"){t.Fatal("inline photograph renderer missing")}
 if !strings.Contains(page,".content figure.lead-photo img"){t.Fatal("full-width lead-photo layout missing")}
}
func TestNewsGoPaletteAndReaderMargins(t *testing.T) {
 page:=ui.Render(appVersion,appBuild)
 for _,tok:=range []string{"--bg:light-dark(#f5f5f0,#0e141a)","--accent:light-dark(#126877,#69c6d3)","max-width:1200px","@media print"}{
  if !strings.Contains(page,tok){t.Errorf("approved visual design missing %q",tok)}
 }
}
