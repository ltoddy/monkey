package main

import (
	"github.com/ltoddy/monkey/options"
	"github.com/ltoddy/monkey/visitor"
)

func main() {
	opt := options.FromArgs()

	config := &visitor.Config{HttpMethod: opt.HttpMethod, FollowRedirect: opt.FollowRedirect, MaxRedirects: 30}
	v := visitor.New(config)
	url_ := visitor.ParseRawUrl(opt.RawUrl)
	v.Visit(url_)
}
