package main

import (
	"fmt"

	"github.com/mcfly722/goPackages/context"
)

type stringsProvider interface {
	init(parentContext context.Context) chan string
}

type workflow struct {
	inputProvider stringsProvider
}

func newWorkflow(inputProvider stringsProvider) *workflow {
	return &workflow{
		inputProvider: inputProvider,
	}
}

// Go ...
func (workflow *workflow) Go(current context.Context) {
	input := workflow.inputProvider.init(current)

loop:
	for {
		select {
		case line := <-input:
			fmt.Println(fmt.Sprintf("[%v]", line))
			break
		case _, opened := <-current.Opened():
			if !opened {
				break loop
			}
		}
	}
}
