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
	"strings"
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
	if !verifier.ValidHttpMethod(config.Method) {
		log.Fatalf("net/http: invalid method %q", config.Method)
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

	request := makeRequest(v.config.Method, url_, "").WithContext(httptrace.WithClientTrace(context.Background(), trance))
	for _, h := range v.config.Headers {
		if key, value := headerToKeyValue(h); key == "" || value == "" {
			v.logger.Println("ignore invalid header: %s", h)
			continue
		} else {
			request.Header.Set(key, value)
		}
	}

	response, err := v.httpclient.Do(request)
	defer func() { _ = response.Body.Close() }()
	if err != nil {
		v.logger.Fatalln("%v", err)
	}

	v.PrintTraceTimes()
	fmt.Println("%s %s", response.Proto, response.Status)
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

func (v *Visitor) _RecordDNSStart(info httptrace.DNSStartInfo) {
	v.DNSStartAt = time.Now()
	v.logger.Println("DNS request is: %v, start at: %v", info.Host, formatTime(v.DNSStartAt))
}

func (v *Visitor) _RecordDNSDone(info httptrace.DNSDoneInfo) {
	v.DNSDoneAt = time.Now()
	for _, addr := range info.Addrs {
		v.logger.Println("DNS lookup: %s", addr.String())
	}
	v.logger.Println("DNS done at: %v.", formatTime(v.DNSDoneAt))
}

func (v *Visitor) _RecordConnectStart(network, addr string) {
	if v.DNSDoneAt.IsZero() {
		v.DNSDoneAt = time.Now()
	}
	v.ConnectStartAt = time.Now()
	v.logger.Println("Connect to %v(%v), at: %v", addr, network, formatTime(v.ConnectStartAt))
}

func (v *Visitor) _RecordConnectDone(network, addr string, err error) {
	if err != nil {
		v.logger.Fatalln("%v", err)
	}

	v.ConnectDoneAt = time.Now()
	v.logger.Println("Connect to %v(%v) done, at: %v", addr, network, formatTime(v.ConnectDoneAt))
}

func (v *Visitor) _RecordGetConn(_ string) {
	v.GetConnAt = time.Now()
}

func (v *Visitor) _RecordGotConn(info httptrace.GotConnInfo) {
	v.GotConnAt = time.Now()
	v.logger.Println("Got connection %v <-> %v, idle %v, at: %v", info.Conn.LocalAddr(), info.Conn.RemoteAddr(), info.IdleTime, formatTime(v.GotConnAt))
}

func (v *Visitor) _RecordTLSHandshakeStart() {
	v.TLSHandshakeStartAt = time.Now()
	if !v.TLSHandshakeStartAt.IsZero() {
		v.logger.Println("TLS handshake start at: %v", formatTime(v.TLSHandshakeStartAt))
	}
}

func (v *Visitor) _RecordTLSHandshakeDone(state tls.ConnectionState, err error) {
	v.TLSHandshakeDoneAt = time.Now()
	if !v.TLSHandshakeDoneAt.IsZero() {
		v.logger.Println("TLS handshake done at: %v", formatTime(v.TLSHandshakeDoneAt))
	}
}

func (v *Visitor) _RecordGot100Continue() {
	v.Got100ContinueAt = time.Now()
	if !v.Got100ContinueAt.IsZero() {
		v.logger.Println("Got 100(continue) status code at: %v", formatTime(v.Got100ContinueAt))
	}
}

func (v *Visitor) _RecordWait100Continue() {
	v.Wait100ContinueAt = time.Now()
	if !v.Wait100ContinueAt.IsZero() {
		v.logger.Println("Wait 100(continue) status code at: %v", formatTime(v.Wait100ContinueAt))
	}
}

func (v *Visitor) _RecordWroteHeaderField(key string, value []string) {
	v.WroteHeaderFieldAt = time.Now()
	v.logger.Println("Wrote header field (%s: %s) at: %v", key, strings.Join(value, ", "), formatTime(v.WroteHeaderFieldAt))
}

func (v *Visitor) _RecordWroteHeaders() {
	v.WroteHeadersAt = time.Now()
	v.logger.Println("Wrote headers done at: %v", formatTime(v.WroteHeadersAt))
}

func (v *Visitor) _RecordWroteRequest(info httptrace.WroteRequestInfo) {
	v.WroteRequestAt = time.Now()
	v.logger.Println("Wrote request at: %v", formatTime(v.WroteRequestAt))
	if info.Err != nil {
		log.Fatalf("Wrote request failed: %v", info.Err)
	}
}

func (v *Visitor) _RecordGotFirstResponseByte() {
	v.GotFirstResponseByteAt = time.Now()
	v.logger.Println("Start receiving response at: %v", formatTime(v.GotFirstResponseByteAt))
}

const httpTemplate = `` +
	`   DNS Lookup   TCP Connection   Server Processing   Content Transfer` + "\n" +
	`[ %s  |     %s  |        %s  |       %s  ]` + "\n" +
	`             |                |                   |                  |` + "\n" +
	`    namelookup:%s      |                   |                  |` + "\n" +
	`                        connect:%s         |                  |` + "\n" +
	`                                      starttransfer:%s        |` + "\n" +
	`                                                                 total:%s` + "\n"

const HttpRequestTemplate = `            DNS Lookup                           TCP Connect                           Server Processing                           Content Transfer
%-15v    %15v |                                    |                                    |                                           |
        <- %-10v ->           |                                    |                                    |
                                   | %-15v    %15v |                                    |
                                   |         <- %-10v ->          |                                   |
                                   |                                    | %-15v    %15v |
                                   |                                    |         <- %-10v ->          |
`

func (v *Visitor) PrintTraceTimes() {
	fmt.Println(
		httpTemplate,
		formatDuration(v.DNSDoneAt.Sub(v.DNSStartAt)),               // dns lookup
		formatDuration(v.GotConnAt.Sub(v.DNSStartAt)),               // tcp connection
		formatDuration(v.GotFirstResponseByteAt.Sub(v.GotConnAt)),   // server processing
		formatDuration(time.Now().Sub(v.GotConnAt)),                 // content transfer
		formatDuration2(v.DNSDoneAt.Sub(v.DNSDoneAt)),               // name lookup
		formatDuration2(v.GotConnAt.Sub(v.DNSStartAt)),              // connect
		formatDuration2(v.GotFirstResponseByteAt.Sub(v.DNSStartAt)), // start transfer
		formatDuration2(time.Now().Sub(v.DNSStartAt)),               // total
	)
	fmt.Println()
}
