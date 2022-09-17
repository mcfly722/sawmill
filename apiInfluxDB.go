package main

import (
	"fmt"
	"log"

	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/context"
	"github.com/mcfly722/goPackages/jsEngine"
)

// InfluxDB ...
type InfluxDB struct {
	context   context.Context
	eventLoop jsEngine.EventLoop
	runtime   *goja.Runtime
}

// InfluxDBConnectionConfig ...
type InfluxDBConnectionConfig struct {
	api *InfluxDB
	url string
}

// InfluxDBConnection ...
type InfluxDBConnection struct {
	api     *InfluxDB
	url     string
	input   chan goja.Value
	current context.Context
}

// Constructor ...
func (influxDB InfluxDB) Constructor(context context.Context, eventLoop jsEngine.EventLoop, runtime *goja.Runtime) {
	runtime.Set("InfluxDB", &InfluxDB{
		context:   context,
		eventLoop: eventLoop,
		runtime:   runtime,
	})
}

// NewConnection ...
func (influxDB *InfluxDB) NewConnection(url string) *InfluxDBConnectionConfig {
	return &InfluxDBConnectionConfig{
		api: influxDB,
		url: url,
	}
}

// Start ...
func (config *InfluxDBConnectionConfig) Start() *InfluxDBConnection {
	connection := &InfluxDBConnection{
		api:   config.api,
		url:   config.url,
		input: make(chan goja.Value, 1024),
	}

	context, err := config.api.context.NewContextFor(connection, connection.url, "InfluxDB Connection")
	if err != nil {
		log.Fatal(err)
	}

	connection.current = context

	return connection
}

func (connection *InfluxDBConnection) getContext() context.Context {
	return connection.current
}

func (connection *InfluxDBConnection) getInput() chan goja.Value {
	return connection.input
}

// Go ...
func (connection *InfluxDBConnection) Go(current context.Context) {
loop:
	for {
		select {
		case object := <-connection.input:
			current.Log(fmt.Sprintf("influx: %v", object))
			break
		case _, opened := <-current.Opened():
			if !opened {
				break loop
			}
		}
	}
}
