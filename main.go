package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/mcfly722/goPackages/context"
)

var (
	logNameFlag *string
)

func main() {
	logNameFlag = flag.String("logName", "log", "JavaScript plugin")

	flag.Parse()

	logName := *logNameFlag

	fmt.Println(fmt.Sprintf("%s", logName))

	rootContext := context.NewRootContext(context.NewConsoleLogDebugger(100, true))

	ctrlC := make(chan os.Signal, 1)

	signal.Notify(ctrlC, os.Interrupt)

	go func() {
		<-ctrlC
		rootContext.Log(2, "CTRL+C signal")
		rootContext.Cancel()
	}()

	reader := newWorkflow(newTailFileReader(logName, 1000))

	rootContext.NewContextFor(reader, "workflow", "workflow")

	rootContext.Wait()

	os.Exit(0)
}
