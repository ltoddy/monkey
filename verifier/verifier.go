package verifier

import (
	"net/http"

	"github.com/ltoddy/monkey/collection/set"
)

func ValidHttpMethod(method string) bool {
	methods := set.NewSetString()
	methods.Add(http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace)
	return methods.Contains(method)
}
