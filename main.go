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
		HttpMethod:     opt.HttpMethod,
		Include:        opt.Include,
		MaxRedirects:   30,
	}
	v := visitor.New(config, logger.New(opt.Verbose))
	url_ := visitor.ParseRawUrl(opt.RawUrl)
	v.Visit(url_)
}
