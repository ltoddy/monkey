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

	flag.StringVar(&options.HttpMethod, "X", http.MethodGet, "HTTP method to use.")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
	}
	options.RawUrl = args[0]

	return options
}
