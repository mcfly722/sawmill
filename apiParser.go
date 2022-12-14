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
	api              *Parser
	handler          *goja.Callable
	queueStringsSize int64
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
		api:              parser,
		handler:          handler,
		queueStringsSize: 256,
	}
}

// SetQueueStringsSize ...
func (string2JSObject *String2JSObjectConfig) SetQueueStringsSize(size int64) *String2JSObjectConfig {
	string2JSObject.queueStringsSize = size
	return string2JSObject
}

// SendTo ...
func (string2JSObject *String2JSObjectConfig) SendTo(receiver InputJSObjectReceiver) *String2JSObject {

	parser := &String2JSObject{
		api:      string2JSObject.api,
		handler:  string2JSObject.handler,
		input:    make(chan string, string2JSObject.queueStringsSize),
		receiver: receiver,
	}

	context, err := receiver.getContext().NewContextFor(parser, "String2JSObject", "String2JSObject")
	if err != nil {
		log.Fatal(err)
	}

	parser.current = context

	return parser
}

func (string2JSObject *String2JSObject) getContext() context.Context {
	return string2JSObject.current
}

func (string2JSObject *String2JSObject) send(object string) {
	defer func() {
		recover()
	}()
	string2JSObject.input <- object
}

// Go ...
func (string2JSObject *String2JSObject) Go(current context.Context) {

	// close input channel, if closing
	current.SetOnBeforeClosing(func(c context.Context) {
		close(string2JSObject.input)
	})

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

			string2JSObject.receiver.send(string2JSObject.api.runtime.ToValue(result))

			break
		case _, opened := <-current.Opened():
			if !opened {
				break loop
			}
		}
	}
}
