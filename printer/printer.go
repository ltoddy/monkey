package printer

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
)

func PrintHeader(header http.Header, include bool) {
	if include {
		var headers []string
		for key, value := range header {
			headers = append(headers, fmt.Sprintf("%s: %s", key, strings.Join(value, ",")))
		}
		sort.Strings(headers)

		fmt.Println(strings.Join(headers, "\n"))
	}
}

func PrintBody(reader io.Reader) {
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatalf("unable to read response body: %v", err)
	}
	// TODO: print by `Content-Type`
	fmt.Println(string(content))
}
