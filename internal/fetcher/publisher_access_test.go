package fetcher

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestPublisherAccessMessage(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden} {
		message, denied := PublisherAccessMessage(&PublisherHTTPError{Host: "example.org", StatusCode: status})
		if !denied || !strings.Contains(message, "Publisher access restricted") || !strings.Contains(message, "Open the original website") {
			t.Fatalf("status %d: %q, %v", status, message, denied)
		}
		if strings.Contains(strings.ToLower(message), "subscription required") {
			t.Fatalf("claimed paywall: %s", message)
		}
	}
	for _, err := range []error{&PublisherHTTPError{Host: "example.org", StatusCode: http.StatusAccepted}, errors.New("connection reset")} {
		if _, denied := PublisherAccessMessage(err); denied {
			t.Fatalf("unexpected restriction classification: %v", err)
		}
	}
}
