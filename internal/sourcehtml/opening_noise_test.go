package sourcehtml

import "testing"

func TestOpeningFigureNoisyScope(t *testing.T) {
	cases := []struct {
		name string
		html string
		want bool
	}{
		{"featured in article header", `<article><header><div class="entry-featured-image"><img src="/lead.jpg"></div></header></article>`, false},
		{"byline inside article header", `<article><header><div class="byline"><img src="/author.jpg"></div></header></article>`, true},
		{"advertisement inside article header", `<article><header><div class="advertisement"><img src="/ad.jpg"></div></header></article>`, true},
		{"ordinary body figure", `<article><figure><img src="/body.jpg"></figure></article>`, false},
		{"navigation in header", `<article><header><nav><img src="/nav.jpg"></nav></header></article>`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := Parse(tc.html)
			scope := Find(root, func(n *Node) bool { return n.Tag == "article" })[0]
			img := Find(scope, func(n *Node) bool { return n.Tag == "img" })[0]
			if got := OpeningFigureNoisy(img, scope); got != tc.want {
				t.Fatalf("OpeningFigureNoisy = %v, want %v", got, tc.want)
			}
		})
	}
	if !OpeningFigureNoisy(nil, nil) { t.Fatal("nil scope must be rejected") }
}
