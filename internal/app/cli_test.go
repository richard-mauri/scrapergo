package app

import (
	"bytes"

	"flag"
	"strings"
	"testing"
)

func TestResolvePort(t *testing.T) {
	for _, tc := range []struct {
		args      []string
		env, want string
		fail      bool
	}{
		{nil, "", "8080", false}, {nil, "9000", "9000", false}, {[]string{"--port", "8081"}, "9000", "8081", false},
		{[]string{"-port=9001"}, "", "9001", false}, {[]string{"--port", "0"}, "", "", true},
		{[]string{"--port", "65536"}, "", "", true}, {[]string{"--port", "foo"}, "", "", true},
	} {
		got, err := ResolvePort(tc.args, tc.env, &bytes.Buffer{})
		if (err != nil) != tc.fail || (!tc.fail && got != tc.want) {
			t.Errorf("%v %q => %q %v", tc.args, tc.env, got, err)
		}
	}
}
func TestHelp(t *testing.T) {
	var out bytes.Buffer
	_, err := ResolvePort([]string{"--help"}, "", &out)
	if err != flag.ErrHelp || !strings.Contains(out.String(), "--port") {
		t.Fatalf("help: %v %q", err, out.String())
	}
}
