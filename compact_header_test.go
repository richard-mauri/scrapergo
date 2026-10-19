package main

import (
	"strings"
	"testing"
)

func TestCompactHeaderRenderer(t *testing.T) {
	for _, needle := range []string{"ScraperGo · v" + appVersion, "identifyingImage", "const first=b.kind==='img'?b:data.blocks[bi+1]", "const head=b.kind==='img'?data.blocks[bi+1]:b", "metadata.length>=1", "className='compact-entry'", "className='compact-info'", "bi+=1+metadata.length"} {
		if !strings.Contains(renderedIndexHTML(), needle) {
			t.Fatalf("missing compact-header behavior: %s", needle)
		}
	}
}
