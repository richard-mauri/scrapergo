package imagedelivery

import (
 "testing"

)

func TestHTTPSImageURL(t *testing.T) {
 cases := []struct{input,want string; valid bool}{
  {"http://static1.squarespace.com/a+b.jpg?format=1500w","https://static1.squarespace.com/a+b.jpg?format=1500w",true},
  {"https://example.org/pic.webp","https://example.org/pic.webp",true},
  {"ftp://example.org/pic.jpg","",false},
  {"javascript:alert(1)","",false},
 }
 for _,tc:=range cases {
  got,err:=HTTPSImageURL(tc.input)
  if tc.valid && (err!=nil||got!=tc.want) { t.Errorf("%s: got %q err %v",tc.input,got,err) }
  if !tc.valid&&err==nil { t.Errorf("%s unexpectedly accepted",tc.input) }
 }
}
