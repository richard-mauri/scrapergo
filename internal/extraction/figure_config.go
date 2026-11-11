package extraction

import (
 "fmt"
 "os"
 "strconv"
 "strings"
)

// ArticleImageCandidateLimitEnv is a process-start setting that bounds the
// number of unique, valid selected-HTML images considered for positioning.
// It does not change which images qualify as editorial photographs.
const ArticleImageCandidateLimitEnv = "SCRAPERGO_ARTICLE_IMAGE_CANDIDATE_LIMIT"

const (
 // DefaultArticleImageCandidateLimit preserves the existing reader behavior.
 DefaultArticleImageCandidateLimit = 12
 // MinArticleImageCandidateLimit rejects zero, negative and disabled limits.
 MinArticleImageCandidateLimit = 1
 // MaxArticleImageCandidateLimit is a fixed resource-safety ceiling, not
 // an editorial-image selection recommendation.
 MaxArticleImageCandidateLimit = 200
)

// FigureSelectionConfig is an immutable, validated extraction configuration.
// Use NewFigureSelectionConfig or DefaultFigureSelectionConfig to construct it.
type FigureSelectionConfig struct {
 maxCandidates int
}

func (c FigureSelectionConfig) ArticleImageCandidateLimit() int {
 return c.maxCandidates
}

func NewFigureSelectionConfig(maxCandidates int) (FigureSelectionConfig,error) {
 if maxCandidates<MinArticleImageCandidateLimit || maxCandidates>MaxArticleImageCandidateLimit {
  return FigureSelectionConfig{},fmt.Errorf("article image candidate limit %d is outside %d..%d",maxCandidates,MinArticleImageCandidateLimit,MaxArticleImageCandidateLimit)
 }
 return FigureSelectionConfig{maxCandidates:maxCandidates},nil
}

func DefaultFigureSelectionConfig() FigureSelectionConfig {
 return FigureSelectionConfig{maxCandidates:DefaultArticleImageCandidateLimit}
}

// FigureSelectionConfigFromEnvironment reads the owner-controlled setting.
// An absent value uses the unchanged default; an invalid/empty override is an
// explicit error rather than a silent fallback. No mutable global state.
func FigureSelectionConfigFromEnvironment() (FigureSelectionConfig,error) {
 raw,ok:=os.LookupEnv(ArticleImageCandidateLimitEnv)
 if !ok {return DefaultFigureSelectionConfig(),nil}
 limit,err:=strconv.Atoi(strings.TrimSpace(raw))
 if err!=nil {return FigureSelectionConfig{},fmt.Errorf("%s must be an integer (%d..%d): %w",ArticleImageCandidateLimitEnv,MinArticleImageCandidateLimit,MaxArticleImageCandidateLimit,err)}
 cfg,err:=NewFigureSelectionConfig(limit)
 if err!=nil {return FigureSelectionConfig{},fmt.Errorf("%s: %w",ArticleImageCandidateLimitEnv,err)}
 return cfg,nil
}
