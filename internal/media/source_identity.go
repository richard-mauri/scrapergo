package media

import (
 "net/url"
 "strings"
)

// imageSourceIdentity identifies renditions only when the CDN supplies
// an explicit original asset URL. It never guesses by filename or dimensions.
func SourceIdentity(raw string) string {
 u,err:=url.Parse(raw)
 if err!=nil || (u.Scheme!="https"&&u.Scheme!="http") || u.Hostname()=="" {return raw}
 if strings.EqualFold(u.Hostname(),"npr.brightspotcdn.com") {
  embedded:=u.Query().Get("url")
  if asset,err:=url.Parse(embedded);err==nil && (asset.Scheme=="http"||asset.Scheme=="https") && asset.Hostname()!="" {
   asset.Fragment=""
   return asset.String()
  }
 }
 u.Fragment=""
 return u.String()
}
