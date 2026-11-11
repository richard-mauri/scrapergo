package diagnostics

import (
 "crypto/hmac"
 "crypto/rand"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "os"
 "path/filepath"
 "strings"
 "sync"
 "time"
 "net/url"

 "clean-article-reader/internal/article"
 "clean-article-reader/internal/extraction"
)

// RecoveryRecorder is a local append-only, opt-in diagnostic sink.
// It never stores full article URLs, HTML or article text.
type RecoveryRecorder struct {
 mu sync.Mutex
 path string
 key []byte
 build string
}

type RecoveryRecord struct {
 Timestamp string `json:"timestamp"`
 Publisher string `json:"publisher"`
 ArticleID string `json:"articleId"`
 Build string `json:"build,omitempty"`
 Method string `json:"method,omitempty"`
 ExtractionError bool `json:"extractionError,omitempty"`
 FinalBlocks int `json:"finalBlocks"`
 FinalWords int `json:"finalWords"`
 Stages []extraction.RecoveryStageResult `json:"stages"`
}

// NewRecoveryRecorder stores data in the diagnostics folder under the process
// working directory. scrapergo.sh always changes to the repository root.
func NewRecoveryRecorder(directory, build string) (*RecoveryRecorder, error) {
 if err := os.MkdirAll(directory,0700); err != nil { return nil, err }
 keyFile := filepath.Join(directory, ".recovery-key")
 key, err := os.ReadFile(keyFile)
 if errors.Is(err,os.ErrNotExist) {
  key=make([]byte,32)
  if _,err=rand.Read(key);err!=nil{return nil,err}
  f,createErr:=os.OpenFile(keyFile,os.O_WRONLY|os.O_CREATE|os.O_EXCL,0600)
  if errors.Is(createErr,os.ErrExist) {
   key,err=os.ReadFile(keyFile)
  } else if createErr!=nil {
   return nil,createErr
  } else {
   _,err=f.Write(key)
   closeErr:=f.Close()
   if err==nil {err=closeErr}
  }
 }
 if err!=nil{return nil,err}
 if len(key)!=32{return nil,fmt.Errorf("invalid diagnostics key length")}
 return &RecoveryRecorder{path:filepath.Join(directory,"recovery.jsonl"),key:key,build:build},nil
}

func (r *RecoveryRecorder) Record(source string, a article.Article, extractionErr error, trace *extraction.RecoveryTrace) error {
 if r==nil||trace==nil{return nil}
 u,err:=url.Parse(source)
 if err!=nil{return err}
 host:=strings.ToLower(u.Hostname())
 if host=="" {return fmt.Errorf("missing publisher host")}
 // HMAC keeps the identifier stable across sessions without publishing
 // any URL or predictable unsalted URL hash.
 h:=hmac.New(sha256.New,r.key)
 _,_=h.Write([]byte(source))
 id:=hex.EncodeToString(h.Sum(nil)[:16])
 record:=RecoveryRecord{
  Timestamp:time.Now().UTC().Format(time.RFC3339Nano),
  Publisher:host,ArticleID:id,Build:r.build,
  Method:a.Method,ExtractionError:extractionErr!=nil,
  FinalBlocks:len(a.Blocks),FinalWords:a.WordCount,
  Stages:trace.Results(),
 }
 payload,err:=json.Marshal(record)
 if err!=nil{return err}
 r.mu.Lock()
 defer r.mu.Unlock()
 f,err:=os.OpenFile(r.path,os.O_WRONLY|os.O_APPEND|os.O_CREATE,0600)
 if err!=nil{return err}
 if _,err=f.Write(append(payload,'\n'));err!=nil{_ = f.Close();return err}
 return f.Close()
}
