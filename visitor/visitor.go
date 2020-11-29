package visitor

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ltoddy/monkey/collection/set"
	"github.com/ltoddy/monkey/options"
)

type Visitor struct {
	opt        *options.Options
	httpclient *http.Client
}

func New(opt *options.Options) *Visitor {
	if !validMethod(opt.HttpMethod) {
		log.Fatalf("net/http: invalid method %q", opt.HttpMethod)
	}

	client := &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       1 * time.Minute,
	}

	return &Visitor{opt: opt, httpclient: client}
}

func (v *Visitor) Visit() {
	request, err := http.NewRequest(v.opt.HttpMethod, v.opt.RawUrl, nil)
	if err != nil {
		log.Fatalf("new request failed: %v", err)
	}

	response, err := v.httpclient.Do(request)
	if err != nil {
		log.Fatalf("fetch failed: %v", err)
	}

	fmt.Printf("%s %s\n", response.Proto, response.Status)
}

func validMethod(method string) bool {
	methods := set.NewSetString()
	methods.Add(http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace)
	return methods.Contains(method)
}
