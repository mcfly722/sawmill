package main

import (
	"log"

	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/context"
	"github.com/mcfly722/goPackages/jsEngine"
)

// Parser ...
type Parser struct {
	context   context.Context
	eventLoop jsEngine.EventLoop
	runtime   *goja.Runtime
}

// String2JSObjectConfig ...
type String2JSObjectConfig struct {
	api     *Parser
	handler *goja.Callable
}

// String2JSObject ...
type String2JSObject struct {
	api      *Parser
	handler  *goja.Callable
	input    chan string
	receiver InputJSObjectReceiver
	current  context.Context
}

// Constructor ...
func (parser Parser) Constructor(context context.Context, eventLoop jsEngine.EventLoop, runtime *goja.Runtime) {
	runtime.Set("Parser", &Parser{
		context:   context,
		eventLoop: eventLoop,
		runtime:   runtime,
	})
}

// NewString2JSObject ...
func (parser *Parser) NewString2JSObject(handler *goja.Callable) *String2JSObjectConfig {
	return &String2JSObjectConfig{
		api:     parser,
		handler: handler,
	}
}

// SendTo ...
func (string2JSObject *String2JSObjectConfig) SendTo(receiver InputJSObjectReceiver) *String2JSObject {

	parser := &String2JSObject{
		api:      string2JSObject.api,
		handler:  string2JSObject.handler,
		input:    make(chan string, 1024),
		receiver: receiver,
	}

	context, err := receiver.getContext().NewContextFor(parser, "String2JSObject", "String2JSObject")
	if err != nil {
		log.Fatal(err)
	}

	parser.current = context

	return parser
}

func (string2JSObject *String2JSObject) getInput() chan string {
	return string2JSObject.input
}

func (string2JSObject *String2JSObject) getContext() context.Context {
	return string2JSObject.current
}

// Go ...
func (string2JSObject *String2JSObject) Go(current context.Context) {
loop:
	for {
		select {
		case line := <-string2JSObject.input:
			jsLine := string2JSObject.api.runtime.ToValue(line)
			result, err := string2JSObject.api.eventLoop.CallHandler(string2JSObject.handler, jsLine)
			if err != nil {
				current.Log(51, err.Error())
				break
			}
			string2JSObject.receiver.getInput() <- string2JSObject.api.runtime.ToValue(result)
			break
		case _, opened := <-current.Opened():
			if !opened {
				break loop
			}
		}
	}
}
