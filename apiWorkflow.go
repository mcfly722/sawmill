package main

import (
	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/context"
)

// InputJSObjectReceiver ...
type InputJSObjectReceiver interface {
	getContext() context.Context
	send(object goja.Value)
}

// InputStringReceiver ...
type InputStringReceiver interface {
	getContext() context.Context
	send(object string)
}
