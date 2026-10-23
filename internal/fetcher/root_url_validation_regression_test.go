package fetcher_test

import (
 "clean-article-reader/internal/fetcher"
 "context"
 "testing"
)

func TestRejectLocal(t *testing.T) {
	for _, u := range []string{"file:///etc/passwd", "http://localhost/private", "http://127.0.0.1/", "http://10.0.0.1/"} {
		if _, err := fetcher.ValidatedURL(context.Background(), u); err == nil {
			t.Errorf("accepted %s", u)
		}
	}
}

