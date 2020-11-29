package visitor

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func ParseRawUrl(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(fmt.Sprintf("invalid url(%s): %v", rawurl, u))
	}
	if u.Scheme == "" {
		u.Scheme = "http"
		if strings.Contains(u.Scheme, ":443") {
			u.Scheme = "https"
		}
	}

	return u
}

func makeRequest(method string, url *url.URL, body string) *http.Request {
	var reader io.Reader = strings.NewReader(body)
	if strings.HasPrefix(body, "@") {
		filename := body[1:]
		f, err := os.Open(filename)
		if err != nil {
			log.Fatalf("failed to open data file %s: %v", filename, err)
		}
		reader = f
	}

	request, err := http.NewRequest(method, url.String(), reader)
	if err != nil {
		log.Fatalf("unable to create request: %v", err)
	}
	return request
}

func isRedirect(response *http.Response) bool {
	return response.StatusCode <= 400 && response.StatusCode >= 300
}
