package main

import (
	"log"
	"net/url"
	"strings"
)

// Diagnostic only: no raw article contents are logged and no extraction changes.
func slateDiagnosticTarget(source string) bool {
	u, err := url.Parse(source)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Hostname(), "arstechnica.com") &&
		strings.Contains(u.Path, "/visiting-slates-first-showroom-as-its-25000-electric-truck-nears-production/")
}

func logSlateRetrievalDiagnostic(raw, source string) {
	if !slateDiagnosticTarget(source) {
		return
	}
	// Raw markers may come from JSON/script content; visible markers cannot.
	visible := strings.ToLower(textOf(parseHTML(raw)))
	lower := strings.ToLower(raw)
	log.Printf("slate_diagnostic stage=download html_bytes=%d raw_slate_garage=%t visible_slate_garage=%t raw_repairpal=%t visible_repairpal=%t raw_closing=%t visible_closing=%t raw_pegboard=%t raw_interior_models=%t",
		len(raw),
		strings.Contains(lower, "beyond the default blank slate body style"),
		strings.Contains(visible, "beyond the default blank slate body style"),
		strings.Contains(lower, "for service, slate will be using"),
		strings.Contains(visible, "for service, slate will be using"),
		strings.Contains(lower, "now it’s time to deliver") || strings.Contains(lower, "now it's time to deliver"),
		strings.Contains(visible, "now it’s time to deliver") || strings.Contains(visible, "now it's time to deliver"),
		strings.Contains(lower, "slate-visit-chicago-pegboard"),
		strings.Contains(lower, "slate-visit-chicago-interior-models"),
	)
}

func logSlateExtractionDiagnostic(blocks []Block, source string) {
	if !slateDiagnosticTarget(source) {
		return
	}
	var texts []string
	images := 0
	for _, b := range blocks {
		texts = append(texts, b.Text)
		if b.Kind == "img" {
			images++
			texts = append(texts, b.URL)
		}
	}
	value := strings.ToLower(strings.Join(texts, " "))
	log.Printf("slate_diagnostic stage=extraction blocks=%d images=%d slate_garage=%t repairpal=%t closing=%t pegboard=%t interior_models=%t",
		len(blocks), images,
		strings.Contains(value, "beyond the default blank slate body style"),
		strings.Contains(value, "for service, slate will be using"),
		strings.Contains(value, "now it’s time to deliver") || strings.Contains(value, "now it's time to deliver"),
		strings.Contains(value, "slate-visit-chicago-pegboard"),
		strings.Contains(value, "slate-visit-chicago-interior-models"),
	)
}
