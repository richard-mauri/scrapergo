package media

import "testing"

func TestOpeningFigureEligibleStructuralEvidence(t *testing.T) {
	cases := []struct {
		name string
		ancestors []Ancestor
		want bool
	}{
		{"article header featured image", []Ancestor{{Tag:"div",Class:"entry-featured-image"}, {Tag:"header",Class:"entry-header"}, {Tag:"article"}}, true},
		{"figure in body", []Ancestor{{Tag:"picture"}, {Tag:"figure"}, {Tag:"article"}}, true},
		{"hero class", []Ancestor{{Tag:"div",Class:"hero-image"}}, true},
		{"ordinary header decoration", []Ancestor{{Tag:"header"}, {Tag:"article"}}, false},
		{"byline picture", []Ancestor{{Tag:"picture"}, {Tag:"div",Class:"byline"}, {Tag:"header"}}, false},
		{"author-featured collision", []Ancestor{{Tag:"div",Class:"featured-image"}, {Tag:"div",Class:"author-bio"}}, false},
		{"navigation figure", []Ancestor{{Tag:"figure"}, {Tag:"nav"}}, false},
		{"unrelated image", []Ancestor{{Tag:"div",Class:"entry-content"}}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name,func(t *testing.T) {
			if got := OpeningFigureEligible(tc.ancestors); got != tc.want {
				t.Fatalf("OpeningFigureEligible(%+v) = %v; want %v", tc.ancestors, got, tc.want)
			}
		})
	}
}
