package imagedelivery

import (
 "context"
 "crypto/tls"
 "errors"
 "net"
 "net/url"
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
		{&url.Error{Op: "Get", Err: &net.DNSError{Err: "no such host", Name: "example.invalid"}}, "dns"},
		{&url.Error{Op: "Get", Err: &tls.CertificateVerificationError{}}, "tls-certificate"},
		{&url.Error{Op: "Get", Err: tls.RecordHeaderError{}}, "tls-handshake"},
		{&url.Error{Op: "Get", Err: &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("refused")}}, "connection"},
	}
	for _, tc := range cases {
		if got := FailureKind(tc.err); got != tc.want {
			t.Errorf("%v: got %q want %q", tc.err, got, tc.want)
		}
	}
}

