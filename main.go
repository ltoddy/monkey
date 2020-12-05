package main

import (
	"github.com/ltoddy/monkey/logger"
	"github.com/ltoddy/monkey/options"
	"github.com/ltoddy/monkey/visitor"
)

func main() {
	opt := options.FromArgs()

	config := &visitor.Config{
		FollowRedirect: opt.FollowRedirect,
		Method:         opt.Method,
		Include:        opt.Include,
		MaxRedirects:   30,
		Headers:        opt.Headers,
	}
	v := visitor.New(config, logger.New(opt.Verbose))
	url_ := visitor.ParseRawUrl(opt.RawUrl)
	v.Visit(url_)
}
