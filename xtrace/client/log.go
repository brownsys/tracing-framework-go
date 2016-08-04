package client

import (
	"fmt"

	"github.com/brownsys/tracing-framework-go/tracingplane/pubsub/client"
	"github.com/golang/protobuf/proto"
)

var defaultLogger *logger

func init() {
	defaultLogger = new(logger)

	var err error
	defaultLogger.c, err = client.NewClient("localhost:5563")
	if err != nil {
		panic(err)
	}
}

type logger struct {
	c *client.Client
}

func Log(str string) { defaultLogger.log(str) }

var topic = []byte("xtrace")

func (l *logger) log(str string) {
	parent, event := newEvent()
	var report XTraceReportv4

	report.ParentEventId = []int64{parent}
	report.EventId = new(int64)
	*report.EventId = event
	report.Label = new(string)
	*report.Label = str

	buf, err := proto.Marshal(&report)
	if err != nil {
		panic(fmt.Errorf("internal error: %v", err))
	}

	l.c.Publish(topic, buf)
}
