package pageclassify

import (
	"fmt"
	"testing"
)

func TestMarketingLandingPageIsNotArticle(t *testing.T) {
	html := `<html><main><h1>Build great things with our platform</h1>`
	for i := 0; i < 6; i++ {
		html += fmt.Sprintf(`<section><h2>Powerful tools for every team, number %d</h2><p>Collaborate with your colleagues and bring your projects to life.</p><a href="/features/%d">Explore features</a></section>`, i, i)
	}
	html += `</main></html>`
	if !IsLanding(html) {
		t.Fatal("product/marketing landing page should not be treated as an article")
	}
}

func TestTechnicalArticleWithHeadingsAndLinksIsNotLanding(t *testing.T) {
	html := `<html><main><h1>How the compiler works</h1>`
	for i := 0; i < 8; i++ {
		html += fmt.Sprintf(`<section><h2>Compiler phase %d explained in detail</h2><p>This technical section provides substantial explanation of a compiler operation, including the data structures it manipulates and the precise reasoning behind each transformation. It also explains the resulting behavior and why programmers should care about the implementation choices.</p><a href="/docs/%d">Read documentation</a></section>`, i, i)
	}
	html += `</main></html>`
	if IsLanding(html) {
		t.Fatal("narrative technical article must not be classified as landing page")
	}
}
