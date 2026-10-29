package bootstrap

import (
 "context"
 "encoding/json"
 "net"
 "net/http"
 "net/http/httptest"
 "strings"
 "time"

 "clean-article-reader/internal/extraction"
 "clean-article-reader/internal/fetcher"
 "clean-article-reader/internal/pageclassify"
)

// classificationTraceHandler is an operator-only, read-only diagnostic. It
// fetches through the same validated downloader as /api/read.
func classificationTraceHandler(read http.Handler) http.HandlerFunc {
 return func(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type","application/json; charset=utf-8")
  w.Header().Set("Cache-Control","no-store")
  if r.Method != http.MethodPost { w.Header().Set("Allow","POST"); http.Error(w,"POST required",http.StatusMethodNotAllowed);return }
  peer,_,err:=net.SplitHostPort(r.RemoteAddr)
  ip:=net.ParseIP(peer)
  if err != nil || ip == nil || !ip.IsLoopback() {http.Error(w,"localhost clients only",http.StatusForbidden);return}
  r.Body=http.MaxBytesReader(w,r.Body,4096)
  var input struct{URL string `json:"url"`}
  if json.NewDecoder(r.Body).Decode(&input)!=nil || strings.TrimSpace(input.URL)=="" || len(input.URL)>2048 {http.Error(w,"invalid URL JSON",http.StatusBadRequest);return}
  ctx,cancel:=context.WithTimeout(r.Context(),15*time.Second);defer cancel()
  if _,err=fetcher.ValidatedURL(ctx,input.URL);err!=nil {http.Error(w,"invalid public URL",http.StatusUnprocessableEntity);return}
  raw,finalURL,err:=fetcher.Download(ctx,input.URL)
  if err!=nil {_=json.NewEncoder(w).Encode(map[string]any{"url":input.URL,"stage":"download","error":err.Error()});return}
  initialIndex:=pageclassify.IsIndex(raw)
  initialLanding:=pageclassify.IsLanding(raw)
  hasPartial:=extraction.HasPartialArticleEvidence(raw,finalURL)
  extracted,extractErr:=ExtractWithEvidence(raw,finalURL)
  wordCount:=0
  if extractErr==nil {wordCount=extracted.Article.WordCount}
  thin:=pageclassify.IsThinPortal(raw,wordCount)
  sparse:=pageclassify.IsSparseRootLanding(raw,finalURL,wordCount)
  search:=pageclassify.IsSearchPortal(raw)
  // Execute the real classification writer against the already-fetched HTML.
  // This prevents a second publisher request and records the actual result.
  recorder:=httptest.NewRecorder()
  WriteExtractedArticle(recorder,raw,finalURL)
  var result map[string]any
  _=json.Unmarshal(recorder.Body.Bytes(),&result)
  outcome:=map[string]any{"status":recorder.Code}
  if kind,ok:=result["kind"];ok {outcome["kind"]=kind}
  if recoderErr,ok:=result["error"];ok {outcome["error"]=recoderErr}
  if recorder.Code==http.StatusOK {outcome["kind"]="article"}
  trace:=map[string]any{
   "url":input.URL,"finalURL":finalURL,"htmlBytes":len(raw),
   "checks":map[string]bool{
    "isIndex":initialIndex,"isLanding":initialLanding,
    "hasPartialArticleEvidence":hasPartial,
    "isThinPortal":thin,"isSparseRootLanding":sparse,"isSearchPortal":search,
   },
   "extraction":map[string]any{"succeeded":extractErr==nil,"wordCount":wordCount},
   "outcome":outcome,
  }
  if extractErr!=nil {trace["extraction"].(map[string]any)["error"]=extractErr.Error()}
  triggered:="none"
  if (initialIndex||initialLanding)&&!hasPartial {triggered="pre_extraction_index_or_landing"} else if ((extractErr==nil&&(thin||sparse))||(extractErr!=nil&&search))&&!hasPartial {triggered="post_extraction_portal"}
  trace["triggeredStage"]=triggered
  _=json.NewEncoder(w).Encode(trace)
 }
}
