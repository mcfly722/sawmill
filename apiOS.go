package main

import (
	"os"

	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/context"
	"github.com/mcfly722/goPackages/jsEngine"
)

// OperatingSystem ...
type OperatingSystem struct {
	context   context.Context
	eventLoop jsEngine.EventLoop
	runtime   *goja.Runtime
}

// Constructor ...
func (operatingSystem OperatingSystem) Constructor(context context.Context, eventLoop jsEngine.EventLoop, runtime *goja.Runtime) {
	runtime.Set("OS", &OperatingSystem{
		context:   context,
		eventLoop: eventLoop,
		runtime:   runtime,
	})
}

// Getenv ...
func (operatingSystem *OperatingSystem) Getenv(name string) string {
	return os.Getenv(name)
}
