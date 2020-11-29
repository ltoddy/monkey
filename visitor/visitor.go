package visitor

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/ltoddy/monkey/options"
	"github.com/ltoddy/monkey/verifier"
)

type Visitor struct {
	opt        *options.Options
	httpclient *http.Client

	DNSStartAt             time.Time // when a DNS lookup begins.
	DNSDoneAt              time.Time // when a DNS lookup ends.
	ConnectStartAt         time.Time // when a new connection's Dial begins.
	ConnectDoneAt          time.Time // when a new connection's Dial completes.
	GetConnAt              time.Time // before a connection is created
	GotConnAt              time.Time // after a successful connection is obtained.
	GotFirstResponseByteAt time.Time // when the first byte of the response headers is available.
	TLSHandshakeStartAt    time.Time // when the TLS handshake is started.
	TLSHandshakeDoneAt     time.Time // after the TLS handshake.
	Got100ContinueAt       time.Time // if the server replies with a "100 Continue" response.
	Wait100ContinueAt      time.Time // if the Request specified "Expect: 100-continue" and the Transport has written the request headers but is waiting for "100 Continue" from the server before writing the request body.
	WroteHeaderFieldAt     time.Time // after the Transport has written each request header.
	WroteHeadersAt         time.Time // after the Transport has written all request headers.
	WroteRequestAt         time.Time
}

func New(opt *options.Options) *Visitor {
	if !verifier.ValidHttpMethod(opt.HttpMethod) {
		log.Fatalf("net/http: invalid method %q", opt.HttpMethod)
	}

	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		TLSHandshakeTimeout:   1 * time.Minute,
		IdleConnTimeout:       1 * time.Minute,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return (&net.Dialer{Timeout: 30 * time.Minute, KeepAlive: 30 * time.Minute}).DialContext(ctx, "tcp4", addr)
		},
	}
	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			// always refuse to follow redirects, visit doest that manually if required
			return http.ErrUseLastResponse
		},
	}

	return &Visitor{opt: opt, httpclient: client}
}

func (v *Visitor) Visit() {
	trance := &httptrace.ClientTrace{
		DNSStart:             v._RecordDNSStart,
		DNSDone:              v._RecordDNSDone,
		ConnectStart:         v._RecordConnectStart,
		ConnectDone:          v._RecordConnectDone,
		GotConn:              v._RecordGotConn,
		GotFirstResponseByte: v._RecordGotFirstResponseByte,
		TLSHandshakeStart:    v._RecordTLSHandshakeStart,
		TLSHandshakeDone:     v._RecordTLSHandshakeDone,
		GetConn:              v._RecordGetConn,
		Got100Continue:       v._RecordGot100Continue,
		Wait100Continue:      v._RecordWait100Continue,
		WroteHeaderField:     v._RecordWroteHeaderField,
		WroteHeaders:         v._RecordWroteHeaders,
		WroteRequest:         v._RecordWroteRequest,
		//PutIdleConn:      nil,
		//Got1xxResponse:   nil,
	}

	u := parseRawUrl(v.opt.RawUrl)
	request := makeRequest(v.opt.HttpMethod, u, "").WithContext(httptrace.WithClientTrace(context.Background(), trance))
	response, err := v.httpclient.Do(request)
	if err != nil {
		log.Fatalf("fetch failed: %v", err)
	}

	fmt.Println(response.Status)
	fmt.Printf("dns start at:               %v\n", v.DNSStartAt)
	fmt.Printf("dns done at:                %v\n", v.DNSDoneAt)
	fmt.Printf("connect start at:           %v\n", v.ConnectStartAt)
	fmt.Printf("connect done at:            %v\n", v.ConnectDoneAt)
	fmt.Printf("got connect at:             %v\n", v.GotConnAt)
	fmt.Printf("got first response byte at: %v\n", v.GotFirstResponseByteAt)
	fmt.Printf("tls hand shark start at:    %v\n", v.TLSHandshakeStartAt)
	fmt.Printf("tls hand shake done at:     %v\n", v.TLSHandshakeDoneAt)
	fmt.Printf("got 100 continue at:        %v\n", v.Got100ContinueAt)
	fmt.Printf("wait 100 continue at:       %v\n", v.Wait100ContinueAt)
	fmt.Printf("wrote header field at:      %v\n", v.WroteHeaderFieldAt)
	fmt.Printf("wrote headers at:           %v\n", v.WroteHeadersAt)
	fmt.Printf("wrote request at:           %v\n", v.WroteRequestAt)
}

func (v *Visitor) _RecordDNSStart(_ httptrace.DNSStartInfo) {
	v.DNSStartAt = time.Now()
}

func (v *Visitor) _RecordDNSDone(_ httptrace.DNSDoneInfo) {
	v.DNSDoneAt = time.Now()
}

func (v *Visitor) _RecordConnectStart(_, _ string) {
	v.ConnectStartAt = time.Now()
}

func (v *Visitor) _RecordConnectDone(network, addr string, err error) {
	if err != nil {
		log.Fatalf("unable to connect to host %v: %v", addr, err)
	}

	v.ConnectDoneAt = time.Now()
	fmt.Printf("Connect to %s\n", addr)
}

func (v *Visitor) _RecordGetConn(_ string) {
	v.GetConnAt = time.Now()
}

func (v *Visitor) _RecordGotConn(_ httptrace.GotConnInfo) {
	v.GotConnAt = time.Now()
}

func (v *Visitor) _RecordGotFirstResponseByte() {
	v.GotFirstResponseByteAt = time.Now()
}

func (v *Visitor) _RecordTLSHandshakeStart() {
	v.TLSHandshakeStartAt = time.Now()
}

func (v *Visitor) _RecordTLSHandshakeDone(_ tls.ConnectionState, _ error) {
	v.TLSHandshakeDoneAt = time.Now()
}

func (v *Visitor) _RecordGot100Continue() {
	v.Got100ContinueAt = time.Now()
}

func (v *Visitor) _RecordWait100Continue() {
	v.Wait100ContinueAt = time.Now()
}

func (v *Visitor) _RecordWroteHeaderField(_ string, _ []string) {
	v.WroteHeaderFieldAt = time.Now()
}

func (v *Visitor) _RecordWroteHeaders() {
	v.WroteHeadersAt = time.Now()
}

func (v *Visitor) _RecordWroteRequest(_ httptrace.WroteRequestInfo) {
	v.WroteRequestAt = time.Now()
}
