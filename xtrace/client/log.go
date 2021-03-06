package client

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/brownsys/tracing-framework-go/xtrace/client/internal"
	"github.com/brownsys/tracing-framework-go/xtrace/internal/pubsub"
	"github.com/golang/protobuf/proto"
)

var client *pubsub.Client

// Connect initializes a connection to the X-Trace
// server. Connect must be called (and must complete
// successfully) before Log can be called.
func Connect(server string) error {
	var err error
	client, err = pubsub.NewClient(server)
	return err
}

var topic = []byte("xtrace")
var processName = strings.Join(os.Args, " ")

// Log logs the given message. Log must not be
// called before Connect has been called successfully.
func Log(str string) {
	if client == nil {
		panic("xtrace/client.Log: no connection to server")
	}

	parent, event := newEvent()
	var report internal.XTraceReportv4

	report.TaskId = new(int64)
	*report.TaskId = GetTaskID()
	report.ParentEventId = []int64{parent}
	report.EventId = new(int64)
	*report.EventId = event
	report.Label = new(string)
	*report.Label = str

	report.Timestamp = new(int64)
	*report.Timestamp = time.Now().UnixNano() / 1000 // milliseconds

	report.ProcessId = new(int32)
	*report.ProcessId = int32(os.Getpid())
	report.ProcessName = new(string)
	*report.ProcessName = processName
	host, err := os.Hostname()
	if err != nil {
		report.Host = new(string)
		*report.Host = host
	}

	buf, err := proto.Marshal(&report)
	if err != nil {
		panic(fmt.Errorf("internal error: %v", err))
	}

	// NOTE(joshlf): Currently, Log blocks until the log message
	// has been written to the TCP connection to the X-Trace server.
	// This makes testing easier, but ideally we should optimize
	// so that the program can block before it quits, but each
	// call to Log is not blocking.
	client.PublishBlock(topic, buf)
}
