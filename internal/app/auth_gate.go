package app

import (
 "context"
 "encoding/json"
 "errors"
 "io"
 "net/http"
 "regexp"
 "strings"
 "time"

 suiteauth "github.com/richard-mauri/homepage/shared/auth"
)

const scrapergoProfileEndpoint="https://home.themauris.org/api/session/profile"
const scrapergoOrigin="https://scrapergo.themauris.org"
var suiteBearerPattern=regexp.MustCompile(`^Bearer [a-f0-9]{64}$`)

// SuiteGate is ScraperGo's mandatory independent per-request identity check.
// The browser may bootstrap public static HTML, but extraction APIs never trust
// browser-supplied role labels, optional activity telemetry or local-mode flags.
type SuiteGate struct {
 ProfileURL string
 Client *http.Client
}

func NewSuiteGate() *SuiteGate {
 return &SuiteGate{
  ProfileURL:scrapergoProfileEndpoint,
  Client:&http.Client{
   Timeout:4*time.Second,
   CheckRedirect:func(*http.Request,[]*http.Request)error{return http.ErrUseLastResponse},
  },
 }
}

func (g *SuiteGate) verify(ctx context.Context,r *http.Request)(suiteauth.Identity,error) {
 authorization:=r.Header.Get("Authorization")
 if !suiteBearerPattern.MatchString(authorization){return suiteauth.Identity{},nil}
 if g==nil||g.ProfileURL==""||g.Client==nil{return suiteauth.Identity{},suiteauth.ErrVerificationUnavailable}
 req,err:=http.NewRequestWithContext(ctx,http.MethodGet,g.ProfileURL,nil)
 if err!=nil{return suiteauth.Identity{},err}
 req.Header.Set("Authorization",authorization)
 req.Header.Set("Origin",scrapergoOrigin)
 req.Header.Set("Accept","application/json")
 response,err:=g.Client.Do(req)
 if err!=nil{return suiteauth.Identity{},err}
 defer response.Body.Close()
 if response.StatusCode==http.StatusUnauthorized||response.StatusCode==http.StatusForbidden{return suiteauth.Identity{},nil}
 if response.StatusCode!=http.StatusOK{return suiteauth.Identity{},suiteauth.ErrVerificationUnavailable}
 var profile struct{
  User string `json:"user"`
  Role suiteauth.Role `json:"role"`
  ExpiresAt string `json:"expiresAt"`
 }
 decoder:=json.NewDecoder(io.LimitReader(response.Body,8192))
 if err:=decoder.Decode(&profile);err!=nil{return suiteauth.Identity{},err}
 expiry,err:=time.Parse(time.RFC3339,profile.ExpiresAt)
 if err!=nil||!time.Now().Before(expiry){return suiteauth.Identity{},nil}
 identity:=suiteauth.Identity{User:strings.TrimSpace(profile.User),Role:profile.Role,Verified:true}
 if !suiteauth.Allowed(identity,suiteauth.UseApplication){return suiteauth.Identity{},nil}
 return identity,nil
}

// Protect wraps the production router. Every /api/* path, including image/PDF,
// classification, diagnostics, presence and future routes is authenticated.
// GET /api/suite/identity supplies an authenticated identity to the bootstrap
// without a second unguarded session authority.
func (g *SuiteGate) Protect(next http.Handler)http.Handler{
 app:=suiteauth.Guard(g.verify,suiteauth.UseApplication,next)
 identity:=suiteauth.Guard(g.verify,suiteauth.UseApplication,http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.Method!=http.MethodGet {
   w.Header().Set("Allow","GET")
   http.Error(w,"Method not allowed",http.StatusMethodNotAllowed);return
  }
  // The shared guard already verified identity. Reuse a scoped independent
  // verification here only for the response identity, never caller labels.
  verified,err:=g.verify(r.Context(),r)
  if err!=nil||!suiteauth.Allowed(verified,suiteauth.UseApplication) {
   http.Error(w,"Homepage verification unavailable",http.StatusServiceUnavailable);return
  }
  w.Header().Set("Content-Type","application/json")
  _=json.NewEncoder(w).Encode(struct{State string `json:"state"`;User string `json:"user"`;Role suiteauth.Role `json:"role"`}{
   State:"signed-in",User:verified.User,Role:verified.Role,
  })
 }))
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.URL.Path=="/api/suite/identity" {identity.ServeHTTP(w,r);return}
  if r.URL.Path=="/api"||strings.HasPrefix(r.URL.Path,"/api/"){app.ServeHTTP(w,r);return}
  next.ServeHTTP(w,r)
 })
}

// Ensure errors.Is remains supported for custom verifiers introduced later.
var _=errors.Is
var _=io.Discard
