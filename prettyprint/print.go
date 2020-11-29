package prettyprint

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
)

func PrettyFormatHeader(header http.Header) string {
	var headers []string
	for key, value := range header {
		headers = append(headers, fmt.Sprintf("%s: %s", key, strings.Join(value, ",")))
	}
	sort.Strings(headers)

	return strings.Join(headers, "\n")
}
