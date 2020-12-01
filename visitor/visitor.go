package visitor

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"time"

	"github.com/ltoddy/monkey/constants"
	"github.com/ltoddy/monkey/logger"
	"github.com/ltoddy/monkey/printer"
	"github.com/ltoddy/monkey/verifier"
)

type Visitor struct {
	config           *Config
	httpclient       *http.Client
	currentRedirects int
	logger           *logger.Logger

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

func New(config *Config, logger *logger.Logger) *Visitor {
	if !verifier.ValidHttpMethod(config.HttpMethod) {
		log.Fatalf("net/http: invalid method %q", config.HttpMethod)
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

	return &Visitor{config: config, httpclient: client, currentRedirects: 0, logger: logger}
}

func (v *Visitor) Visit(url_ *url.URL) {
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

	request := makeRequest(v.config.HttpMethod, url_, "").WithContext(httptrace.WithClientTrace(context.Background(), trance))
	for _, h := range v.config.Headers {
		if key, value := headerToKeyValue(h); key == "" || value == "" {
			v.logger.Printf("ignore invalid header: %s\n", h)
			continue
		} else {
			request.Header.Set(key, value)
		}
	}

	response, err := v.httpclient.Do(request)
	defer func() { _ = response.Body.Close() }()
	if err != nil {
		log.Fatalf("fetch failed: %v", err)
	}

	fmt.Printf("%s %s\n", response.Proto, response.Status)
	fmt.Println()
	printer.PrintHeader(response.Header, v.config.Include)
	printer.PrintBody(response.Body)

	if isRedirect(response) && v.config.FollowRedirect {
		u, err := response.Location()
		if err != nil {
			if err == http.ErrNoLocation {
				return
			}
			log.Fatalf("unable to follow redirect: %v", err)
		}

		v.currentRedirects += 1
		if v.currentRedirects > constants.MaxRedirects {
			log.Fatalf("maximum number of redirects (%d) followed", constants.MaxRedirects)
		}

		v.Visit(u)
	}
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
