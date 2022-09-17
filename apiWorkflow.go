package main

import (
	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/context"
)

// InputJSObjectReceiver ...
type InputJSObjectReceiver interface {
	getContext() context.Context
	getInput() chan goja.Value
}

// InputStringReceiver ...
type InputStringReceiver interface {
	getContext() context.Context
	getInput() chan string
}
