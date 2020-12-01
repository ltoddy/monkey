package options

import (
	"flag"
	"net/http"
)

type Options struct {
	Verbose        bool
	RawUrl         string
	HttpMethod     string
	Include        bool
	FollowRedirect bool
	Headers        []string
}

func FromArgs() *Options {
	options := new(Options)

	flag.BoolVar(&options.Verbose, "v", false, "Make the operation more talkative")
	flag.StringVar(&options.HttpMethod, "X", http.MethodGet, "HTTP method to use")
	flag.BoolVar(&options.Include, "i", false, "Include protocol response headers in the output")
	flag.BoolVar(&options.FollowRedirect, "L", false, "Follow redirects")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
	}
	options.RawUrl = args[0]

	return options
}
