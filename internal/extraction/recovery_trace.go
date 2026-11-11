package extraction

import (
 "bytes"
 "encoding/json"

 "clean-article-reader/internal/article"
)

// RecoveryStage names are deliberately stable across builds for cumulative
// diagnostics. An unvisited stage is reported as not_reached.
var RecoveryStages = []string{
 "readability-render-text", "structured-document", "leading-paragraphs", "source-tables", "caption-reconciliation",
 "selected-html-links", "position-figures", "editorial-continuity",
 "institutional-tail", "heading-anchored-prose", "readability-structure",
 "reader-postprocessing", "legacy-fallback", "partial-html-fallback",
}

type RecoveryStageResult struct {
 Stage string `json:"stage"`
 Outcome string `json:"outcome"`
 BeforeBlocks int `json:"beforeBlocks,omitempty"`
 AfterBlocks int `json:"afterBlocks,omitempty"`
 BeforeHeadings int `json:"beforeHeadings,omitempty"`
 AfterHeadings int `json:"afterHeadings,omitempty"`
 BeforeImages int `json:"beforeImages,omitempty"`
 AfterImages int `json:"afterImages,omitempty"`
 BeforeExactPairs int `json:"beforeExactPairs,omitempty"`
 AfterExactPairs int `json:"afterExactPairs,omitempty"`
 BeforeOverlapPairs int `json:"beforeOverlapPairs,omitempty"`
 AfterOverlapPairs int `json:"afterOverlapPairs,omitempty"`
}

type RecoveryTrace struct {
 results map[string]RecoveryStageResult
}

// NewRecoveryTrace is allocated only for explicitly requested diagnostics.
// It does not retain source HTML, URLs, or article text.
func NewRecoveryTrace() *RecoveryTrace {
 return &RecoveryTrace{results: make(map[string]RecoveryStageResult)}
}

func countRecoveryKinds(blocks []article.Block) (headings, images int) {
 for _, b := range blocks {
  switch b.Kind {
  case "h2", "h3", "h4": headings++
  case "img": images++
  }
 }
 return
}

func (t *RecoveryTrace) mark(stage, outcome string, before, after []article.Block, changed bool) {
 if t == nil { return }
 bh, bi := countRecoveryKinds(before)
 ah, ai := countRecoveryKinds(after)
 be,bo := duplicateCandidatePairs(before)
 ae,ao := duplicateCandidatePairs(after)
 if changed && outcome == "" { outcome = "modified" }
 if outcome == "" { outcome = "executed_unchanged" }
 t.results[stage] = RecoveryStageResult{
  Stage:stage, Outcome:outcome, BeforeBlocks:len(before), AfterBlocks:len(after),
  BeforeHeadings:bh, AfterHeadings:ah, BeforeImages:bi, AfterImages:ai,
  BeforeExactPairs:be, AfterExactPairs:ae, BeforeOverlapPairs:bo, AfterOverlapPairs:ao,
 }
}

// Run observes a stage before and after it executes. JSON snapshots are only
// held temporarily to detect in-place changes to links/parts/blocks.
func (t *RecoveryTrace) Run(stage string, blocks []article.Block, fn func() []article.Block) []article.Block {
 if t == nil { return fn() }
 before, _ := json.Marshal(blocks)
 beforeExact,beforeOverlaps := duplicateCandidatePairs(blocks)
 beforeBlocks := len(blocks)
 bh, bi := countRecoveryKinds(blocks)
 result := fn()
 after, _ := json.Marshal(result)
 afterExact,afterOverlaps := duplicateCandidatePairs(result)
 ah, ai := countRecoveryKinds(result)
 outcome := "executed_unchanged"
 if !bytes.Equal(before, after) { outcome = "modified" }
 t.results[stage] = RecoveryStageResult{
  Stage:stage, Outcome:outcome, BeforeBlocks:beforeBlocks, AfterBlocks:len(result),
  BeforeHeadings:bh, AfterHeadings:ah, BeforeImages:bi, AfterImages:ai,
  BeforeExactPairs:beforeExact, AfterExactPairs:afterExact,
  BeforeOverlapPairs:beforeOverlaps, AfterOverlapPairs:afterOverlaps,
 }
 return result
}

// Observe captures a diagnostic baseline or postprocessing result without
// retaining article text. It has no effect when tracing is disabled.
func (t *RecoveryTrace) Observe(stage string, blocks []article.Block) {
 if t == nil {return}
 t.mark(stage,"observed",blocks,blocks,false)
}

// ObserveTransition records before/after diagnostic counts for a stage
// executed outside the Readability extractor, such as the reader pipeline.
func (t *RecoveryTrace) ObserveTransition(stage string, before, after []article.Block) {
 if t == nil {return}
 t.mark(stage,"observed",before,after,false)
}

func (t *RecoveryTrace) SetOutcome(stage, outcome string) {
 if t == nil { return }
 t.results[stage] = RecoveryStageResult{Stage:stage,Outcome:outcome}
}

func (t *RecoveryTrace) Results() []RecoveryStageResult {
 if t == nil { return nil }
 out := make([]RecoveryStageResult,0,len(RecoveryStages))
 for _, stage := range RecoveryStages {
  if v, ok := t.results[stage]; ok { out=append(out,v) } else {
   out=append(out, RecoveryStageResult{Stage:stage,Outcome:"not_reached"})
  }
 }
 return out
}
