package main

import (
	originalContext "context"
	"fmt"
	"log"
	"time"

	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/context"
	"github.com/mcfly722/goPackages/jsEngine"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdb2api "github.com/influxdata/influxdb-client-go/v2/api"
	influxdb2write "github.com/influxdata/influxdb-client-go/v2/api/write"
)

// InfluxDB ...
type InfluxDB struct {
	context   context.Context
	eventLoop jsEngine.EventLoop
	runtime   *goja.Runtime
}

// JSInfluxDBPoint ...
type JSInfluxDBPoint struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]interface{}
	Timestamp   goja.Value
}

// InfluxDBConnectionConfig ...
type InfluxDBConnectionConfig struct {
	api            *InfluxDB
	url            string
	token          string
	org            string
	bucket         string
	maxBatchSize   int
	sendTimeoutMS  int64
	sendIntervalMS int64
	onSendError    *goja.Callable
	onSendSuccess  *goja.Callable
}

// InfluxDBConnection ...
type InfluxDBConnection struct {
	api     *InfluxDB
	input   chan goja.Value
	current context.Context

	sendTimeoutMS  int64
	sendIntervalMS int64
	maxBatchSize   int

	onSendError   *goja.Callable
	onSendSuccess *goja.Callable

	influxClient   influxdb2.Client
	influxWriteAPI influxdb2api.WriteAPIBlocking
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
		api:            influxDB,
		url:            url,
		token:          "",
		org:            "",
		bucket:         "",
		maxBatchSize:   256,
		sendTimeoutMS:  2000,
		sendIntervalMS: 3000,
	}
}

// SetAuthByToken ...
func (influxDBConnectionConfig *InfluxDBConnectionConfig) SetAuthByToken(token string) *InfluxDBConnectionConfig {
	influxDBConnectionConfig.token = token
	return influxDBConnectionConfig
}

// SetOrganization ...
func (influxDBConnectionConfig *InfluxDBConnectionConfig) SetOrganization(org string) *InfluxDBConnectionConfig {
	influxDBConnectionConfig.org = org
	return influxDBConnectionConfig
}

// SetBucket ...
func (influxDBConnectionConfig *InfluxDBConnectionConfig) SetBucket(bucket string) *InfluxDBConnectionConfig {
	influxDBConnectionConfig.bucket = bucket
	return influxDBConnectionConfig
}

// SetSendMaxBatchSize ...
func (influxDBConnectionConfig *InfluxDBConnectionConfig) SetSendMaxBatchSize(size int) *InfluxDBConnectionConfig {
	influxDBConnectionConfig.maxBatchSize = size
	return influxDBConnectionConfig
}

// SetSendTimeoutMS ...
func (influxDBConnectionConfig *InfluxDBConnectionConfig) SetSendTimeoutMS(timeoutMS int64) *InfluxDBConnectionConfig {
	influxDBConnectionConfig.sendTimeoutMS = timeoutMS
	return influxDBConnectionConfig
}

// SetSendIntervalMS ...
func (influxDBConnectionConfig *InfluxDBConnectionConfig) SetSendIntervalMS(intervalMS int64) *InfluxDBConnectionConfig {
	influxDBConnectionConfig.sendIntervalMS = intervalMS
	return influxDBConnectionConfig
}

// OnSendError ...
func (influxDBConnectionConfig *InfluxDBConnectionConfig) OnSendError(handler *goja.Callable) *InfluxDBConnectionConfig {
	influxDBConnectionConfig.onSendError = handler
	return influxDBConnectionConfig
}

// OnSendSuccess ...
func (influxDBConnectionConfig *InfluxDBConnectionConfig) OnSendSuccess(handler *goja.Callable) *InfluxDBConnectionConfig {
	influxDBConnectionConfig.onSendSuccess = handler
	return influxDBConnectionConfig
}

// Start ...
func (influxDBConnectionConfig *InfluxDBConnectionConfig) Start() *InfluxDBConnection {

	client := influxdb2.NewClient(influxDBConnectionConfig.url, influxDBConnectionConfig.token)
	writeAPI := client.WriteAPIBlocking(influxDBConnectionConfig.org, influxDBConnectionConfig.bucket)

	connection := &InfluxDBConnection{
		api:            influxDBConnectionConfig.api,
		sendTimeoutMS:  influxDBConnectionConfig.sendTimeoutMS,
		sendIntervalMS: influxDBConnectionConfig.sendIntervalMS,
		maxBatchSize:   influxDBConnectionConfig.maxBatchSize,
		onSendError:    influxDBConnectionConfig.onSendError,
		onSendSuccess:  influxDBConnectionConfig.onSendSuccess,
		influxClient:   client,
		influxWriteAPI: writeAPI,
		input:          make(chan goja.Value, 1024),
	}

	context, err := influxDBConnectionConfig.api.context.NewContextFor(connection, influxDBConnectionConfig.url, "InfluxDB Connection")
	if err != nil {
		log.Fatal(err)
	}

	connection.current = context

	return connection
}

func (connection *InfluxDBConnection) getContext() context.Context {
	return connection.current
}

func (connection *InfluxDBConnection) send(object goja.Value) {
	defer func() {
		recover()
	}()
	connection.input <- object
}

func sendBatch(writeAPI influxdb2api.WriteAPIBlocking, timeoutMS int64, batch []*influxdb2write.Point) error {

	if len(batch) == 0 {
		return nil
	}

	ctx, cancel := originalContext.WithTimeout(originalContext.Background(), time.Duration(timeoutMS)*time.Millisecond)
	defer cancel()

	err := writeAPI.WritePoint(ctx, batch...)

	return err
}

func timeFromMsec(msec int64) time.Time {
	sec := msec / 1000
	nsec := (msec % 1000) * 1e6
	return time.Unix(sec, nsec)
}

func jsObject2Time(object goja.Value) (result time.Time, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%v", p)
		}
	}()
	timestamp := timeFromMsec(object.Export().(int64))
	return timestamp, err
}

func jsObject2Point(runtime *goja.Runtime, object goja.Value) (*influxdb2write.Point, error) {

	point := &JSInfluxDBPoint{}
	err := runtime.ExportTo(object, &point)
	if err != nil {
		return nil, err
	}

	timestamp, err := jsObject2Time(point.Timestamp)
	if err != nil {
		//fmt.Println(fmt.Sprintf("TIME stamp error: %v", err))
		return nil, err
	}

	//fmt.Println(fmt.Sprintf("TIMESTAMP: %v", timestamp))

	return influxdb2.NewPoint(
		point.Measurement,
		point.Tags,
		point.Fields,
		timestamp), nil
}

// Go ...
func (connection *InfluxDBConnection) Go(current context.Context) {
	batch := []*influxdb2write.Point{}

	current.SetOnBeforeClosing(func(c context.Context) {
		close(connection.input)
	})

loop:
	for {

	collectBatch:
		for {
			select {
			case <-time.After(time.Duration(connection.sendIntervalMS) * time.Millisecond):
				break collectBatch
			case object := <-connection.input:
				if object != goja.Undefined() {
					point, err := jsObject2Point(connection.api.runtime, object)
					if err != nil {
						current.Log(err)
						break
					}

					batch = append(batch, point)

					if len(batch) >= connection.maxBatchSize {
						break collectBatch
					}

				}
				break
			case _, opened := <-current.Opened():
				if !opened {
					break loop
				}
			}
		}

	sendBatch:
		for len(batch) > 0 {
			select {
			case <-time.After(time.Duration(connection.sendIntervalMS) * time.Millisecond):
				err := sendBatch(connection.influxWriteAPI, connection.sendTimeoutMS, batch)
				if err != nil {
					if connection.onSendError != nil {
						connection.api.eventLoop.CallHandler(connection.onSendError, connection.api.runtime.ToValue(err), connection.api.runtime.ToValue(batch))
					}
				} else {

					if connection.onSendSuccess != nil {
						connection.api.eventLoop.CallHandler(connection.onSendSuccess, connection.api.runtime.ToValue(batch))
					}

					batch = []*influxdb2write.Point{}
					break sendBatch
				}
			case _, opened := <-current.Opened():
				if !opened {
					break loop
				}
			}
		}
	}

	connection.influxClient.Close()
}
