package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestImageRetrievalFailureKind(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{errors.New("upstream image status 403"), "upstream-status"},
		{errors.New("image too large"), "size-limit"},
		{errors.New("unsupported or nonimage response"), "unsupported-source"},
		{context.DeadlineExceeded, "timeout"},
		{errors.New("transport failed"), "image-fetch"},
	}
	for _, tc := range cases {
		if got := imageRetrievalFailureKind(tc.err); got != tc.want {
			t.Errorf("%v: got %q want %q", tc.err, got, tc.want)
		}
	}
}

func TestThumbnailMethodRejectionStillWorks(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnail", strings.NewReader(""))
	w := httptest.NewRecorder()
	articleThumbnail(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status %d", w.Code)
	}
}
