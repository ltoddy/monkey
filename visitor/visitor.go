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
)

type Visitor struct {
	opt        *options.Options
	httpclient *http.Client

	DNSStartAt           time.Time
	DNSDoneAt            time.Time
	ConnectStartAt       time.Time
	ConnectDoneAt        time.Time
	GotConnectAt         time.Time
	GotFirstResponseByte time.Time
	TLSHandshakeStartAt  time.Time
	TLSHandshakeDoneAt   time.Time
}

func New(opt *options.Options) *Visitor {
	if !validMethod(opt.HttpMethod) {
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
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// always refuse to follow redirects
			// visit doest that manually if required
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
		ConnectDone:          func(network, addr string, err error) {},
		GotConn:              v._RecordGotConn,
		GotFirstResponseByte: v._RecordGotFirstResponseByte,
		TLSHandshakeStart:    v._RecordTLSHandshakeStart,
		TLSHandshakeDone:     v._RecordTLSHandshakeDone,
		//GetConn:          nil,
		//PutIdleConn:      nil,
		//Got100Continue:   nil,
		//Got1xxResponse:   nil,
		//WroteHeaderField: nil,
		//WroteHeaders:     nil,
		//Wait100Continue:  nil,
		//WroteRequest:     nil,
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
	fmt.Printf("got connect at:             %v\n", v.GotConnectAt)
	fmt.Printf("got first response byte at: %v\n", v.GotFirstResponseByte)
	fmt.Printf("tls hand shark start at:    %v\n", v.TLSHandshakeStartAt)
	fmt.Printf("tls hand shake done at:     %v\n", v.TLSHandshakeDoneAt)
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
	fmt.Printf("Connect to %s", addr)
}

func (v *Visitor) _RecordGotConn(_ httptrace.GotConnInfo) {
	v.GotConnectAt = time.Now()
}

func (v *Visitor) _RecordGotFirstResponseByte() {
	v.GotFirstResponseByte = time.Now()
}

func (v *Visitor) _RecordTLSHandshakeStart() {
	v.TLSHandshakeStartAt = time.Now()
}

func (v *Visitor) _RecordTLSHandshakeDone(_ tls.ConnectionState, _ error) {
	v.TLSHandshakeDoneAt = time.Now()
}
