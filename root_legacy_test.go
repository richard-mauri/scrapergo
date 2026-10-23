package main

import (
 "context"
 "net/http"
 "net/url"
 "regexp"
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/bootstrap"
 "clean-article-reader/internal/extraction"
 "clean-article-reader/internal/fetcher"
 "clean-article-reader/internal/sourcehtml"
)

type InlineLink = article.InlineLink
type TextPart = article.TextPart
type Block = article.Block
type Article = article.Article
type Node = sourcehtml.Node
var blockedTags = sourcehtml.BlockedTags
var wsRE = regexp.MustCompile(`\s+`)
const maxHTML = 5 << 20
const sampleURL = "https://247wallst.com/investing/2026/10/16/taiwan-is-planning-4-billion-industrial-parks-in-texas-for-nvidias-suppliers/"
func normalize(s string) string {return sourcehtml.Normalize(s)}
func parseHTML(s string) *Node {return sourcehtml.Parse(s)}
func find(n *Node,f func(*Node)bool) []*Node {return sourcehtml.Find(n,f)}
func textOf(n *Node)string{return sourcehtml.Text(n)}
func meta(n *Node,k ...string)string{return sourcehtml.Meta(n,k...)}
func articleImageURL(n *Node, base string)string{return sourcehtml.ImageURL(n,base)}
func safeHost(ctx context.Context,host string)error{return fetcher.SafeHost(ctx,host)}
func validatedURL(ctx context.Context,s string)(*url.URL,error){return fetcher.ValidatedURL(ctx,s)}
func setBrowserHeaders(r *http.Request){fetcher.SetBrowserHeaders(r)}
func setReadeckLikeHeaders(r *http.Request){fetcher.SetReadeckLikeHeaders(r)}
func alternateHeaderRetry(c int)bool{return fetcher.AlternateHeaderRetry(c)}
func download(ctx context.Context,s string)(string,string,error){return fetcher.Download(ctx,s)}
type publisherHTTPError = fetcher.PublisherHTTPError
func publisherAccessMessage(e error)(string,bool){return fetcher.PublisherAccessMessage(e)}
func api(w http.ResponseWriter,r *http.Request){bootstrap.NewArticleReadHandler().ServeHTTP(w,r)}
func newArticleReadHandler()http.Handler{return bootstrap.NewArticleReadHandler()}
func newArticleMetadataHandler(h http.Handler)http.Handler{return bootstrap.NewArticleMetadataHandler(h)}
func writeExtractedArticle(w http.ResponseWriter,raw,src string){bootstrap.WriteExtractedArticle(w,raw,src)}
func extractWithEvidence(raw,src string)(extraction.Result,error){return bootstrap.ExtractWithEvidence(raw,src)}
func slateDiagnosticTarget(src string)bool{return bootstrap.SlateDiagnosticTarget(src)}
func logSlateRetrievalDiagnostic(raw,src string){bootstrap.LogSlateRetrievalDiagnostic(raw,src)}
func logSlateExtractionDiagnostic(b []Block,src string){bootstrap.LogSlateExtractionDiagnostic(b,src)}
