package main

import (
	"github.com/ltoddy/monkey/options"
	"github.com/ltoddy/monkey/visitor"
)

func main() {
	opt := options.FromArgs()
	v := visitor.New(opt)
	v.Visit()
}
