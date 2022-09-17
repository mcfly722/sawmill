package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/context"
	"github.com/mcfly722/goPackages/jsEngine"
)

var (
	pluginFileName string
	pluginBody     string
)

func main() {
	{ // obtain input parameters
		pluginFlag := flag.String("plugin", "plugin.js", "JavaScript plugin")

		flag.Parse()

		pluginFileName = *pluginFlag

		body, err := os.ReadFile(pluginFileName)
		if err != nil {
			log.Fatal(err)
		}
		pluginBody = string(body)
	}

	rootContext := context.NewRootContext(context.NewConsoleLogDebugger(100, true))

	ctrlC := make(chan os.Signal, 1)

	signal.Notify(ctrlC, os.Interrupt)

	go func() {
		<-ctrlC
		rootContext.Log(2, "CTRL+C signal")
		rootContext.Cancel()
	}()

	{ // starting JavaScript plugin EventLoop
		scripts := []jsEngine.Script{jsEngine.NewScript(pluginFileName, string(pluginBody))}
		eventLoop := jsEngine.NewEventLoop(goja.New(), scripts)

		eventLoop.Import(jsEngine.Console{})
		eventLoop.Import(jsEngine.Scheduler{})
		eventLoop.Import(InfluxDB{})
		eventLoop.Import(Parser{})
		eventLoop.Import(FilesTails{})

		_, err := rootContext.NewContextFor(eventLoop, pluginFileName, "eventLoop")
		if err != nil {
			log.Fatal(err)
		}
	}

	rootContext.Wait()

	os.Exit(0)
}
