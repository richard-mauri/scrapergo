package accesspolicy

import "testing"

func TestPublisherRestrictionEvidenceContract(t *testing.T) {
 cases:=[]struct{name,html,kind,evidence string}{
  {"open article","<article><p>Public reporting without restrictions.</p></article>","","unidentified"},
  {"free registration","<article>Register for free to continue reading</article>","registration","free-registration"},
  {"explicit paywall","<article>Paid subscription required</article>","paywall","explicit-paid-subscription"},
  {"structured paywall",`<script type="application/ld+json">{"isAccessibleForFree":false}</script>`,"restricted","structured-paywall-metadata"},
  {"javascript not evidence",`<script>const message="paid subscription required";</script><article>Free content</article>`,"","unidentified"},
  {"css not evidence",`<style>.foo::after{content:"paid subscription required"}</style><article>Public</article>`,"","unidentified"},
 }
 for _,tc:=range cases {t.Run(tc.name,func(t *testing.T){
  if got:=Restriction(tc.html);got!=tc.kind {t.Errorf("kind=%q want %q",got,tc.kind)}
  if got:=Evidence(tc.html);got!=tc.evidence {t.Errorf("evidence=%q want %q",got,tc.evidence)}
 })}
}

func TestRestrictionNoticeIsPreviewAware(t *testing.T) {
 for _,kind:=range []string{"registration","paywall","restricted"}{
  if Notice(kind)=="" {t.Errorf("empty notice for %q",kind)}
 }
}

func TestHostnamePreservesURLHost(t *testing.T) {
 if got:=Hostname("https://www.example.com/story");got!="www.example.com"{t.Fatalf("hostname %q",got)}
}
