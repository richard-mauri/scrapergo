package bootstrap_test

import (
 "testing"
 "clean-article-reader/internal/bootstrap"
)

func TestSlateDiagnosticTarget(t *testing.T) {
	cases := []struct {
		url  string
		want bool
	}{
		{"https://arstechnica.com/cars/2026/10/visiting-slates-first-showroom-as-its-25000-electric-truck-nears-production/", true},
		{"https://www.arstechnica.com/cars/2026/10/visiting-slates-first-showroom-as-its-25000-electric-truck-nears-production/", false},
		{"https://arstechnica.com/other-article/", false},
		{"https://example.com/cars/2026/10/visiting-slates-first-showroom-as-its-25000-electric-truck-nears-production/", false},
	}
	for _, tc := range cases {
		if got := bootstrap.SlateDiagnosticTarget(tc.url); got != tc.want {
			t.Errorf("slateDiagnosticTarget(%q) = %t, want %t", tc.url, got, tc.want)
		}
	}
}
