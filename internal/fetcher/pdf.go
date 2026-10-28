package fetcher

import (
 "bytes"
 "context"
 "errors"
 "fmt"
 "io"
 "net"
 "net/http"
 "net/http/cookiejar"
 "time"
)

// DownloadPDF retrieves an explicitly discovered public PDF. It does not use
// the HTML cache; only successful PDF bytes are returned.
func DownloadPDF(ctx context.Context, address string) ([]byte, string, error) {
 u, err := ValidatedURL(ctx,address)
 if err != nil { return nil,"",err }
 transport := http.DefaultTransport.(*http.Transport).Clone()
 transport.Proxy = nil
 transport.DialContext = func(ctx context.Context, network, addr string)(net.Conn,error){
  host,port,err := net.SplitHostPort(addr)
  if err != nil {return nil,err}
  if err=SafeHost(ctx,host);err!=nil {return nil,err}
  return (&net.Dialer{Timeout:8*time.Second}).DialContext(ctx,network,net.JoinHostPort(host,port))
 }
 jar,_ := cookiejar.New(nil)
 client := &http.Client{
  Timeout:18*time.Second,Transport:transport,Jar:jar,
  CheckRedirect:func(req *http.Request,via []*http.Request)error{
   if len(via)>=5 {return errors.New("too many PDF redirects")}
   _,err:=ValidatedURL(req.Context(),req.URL.String())
   return err
  },
 }
 req,err:=http.NewRequestWithContext(ctx,http.MethodGet,u.String(),nil)
 if err!=nil{return nil,"",err}
 SetBrowserHeaders(req)
 req.Header.Set("Accept","application/pdf,*/*;q=0.8")
 res,err:=client.Do(req)
 if err!=nil{return nil,"",err}
 defer res.Body.Close()
 if res.StatusCode!=http.StatusOK {
  return nil,"",&PublisherHTTPError{Host:res.Request.URL.Hostname(),StatusCode:res.StatusCode}
 }
 const maxPDFBytes = 12 << 20
 if res.ContentLength>maxPDFBytes {return nil,"",fmt.Errorf("PDF exceeds %d bytes",maxPDFBytes)}
 data,err:=io.ReadAll(io.LimitReader(res.Body,maxPDFBytes+1))
 if err!=nil{return nil,"",err}
 if len(data)>maxPDFBytes{return nil,"",fmt.Errorf("PDF exceeds %d bytes",maxPDFBytes)}
 if !bytes.HasPrefix(data,[]byte("%PDF-")) {return nil,"",errors.New("PDF URL did not return PDF content")}
 return data,res.Request.URL.String(),nil
}
