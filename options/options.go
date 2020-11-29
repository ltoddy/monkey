package options

import (
	"flag"
	"net/http"
)

type Options struct {
	RawUrl     string
	HttpMethod string
}

func FromArgs() *Options {
	options := new(Options)

	flag.StringVar(&options.RawUrl, "url", "", "")
	flag.StringVar(&options.HttpMethod, "X", http.MethodGet, "HTTP method to use.")

	flag.Parse()
	return options
}
