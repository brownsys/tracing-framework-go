package client

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/brownsys/tracing-framework-go/local"
)

var token local.Token

type localStorage struct {
	taskID  int64
	eventID int64
}

func init() {
	token = local.Register(&localStorage{
		taskID:  randInt64(),
		eventID: randInt64(),
	}, local.Callbacks{
		func(l interface{}) interface{} {
			// deep copy l
			n := *(l.(*localStorage))
			return &n
		},
	})
}

func getLocal() *localStorage {
	return local.GetLocal(token).(*localStorage)
}

// SetEventID sets the current goroutine's X-Trace Event ID.
// This should be used when propagating Event IDs over RPC
// calls or other channels.
//
// WARNING: This will overwrite any previous Event ID,
// so call with caution.
func SetEventID(eventID int64) {
	getLocal().eventID = eventID
}

// SetTaskID sets the current goroutine's X-Trace Task ID.
// This should be used when propagating Task IDs over RPC
// calls or other channels.
//
// WARNING: This will overwrite any previous Task ID,
// so call with caution.
func SetTaskID(taskID int64) {
	getLocal().taskID = taskID
}

// GetEventID gets the current goroutine's X-Trace Event ID.
// Note that if one has not been set yet, GetEventID will
// return 0. This should be used when propagating Event IDs
// over RPC calls or other channels.
func GetEventID() (eventID int64) {
	return getLocal().eventID
}

// GetTaskID gets the current goroutine's X-Trace Task ID.
// Note that if one has not been set yet, GetTaskID will
// return 0. This should be used when propagating Task IDs
// over RPC calls or other channels.
func GetTaskID() (taskID int64) {
	return getLocal().taskID
}

func newEvent() (parent, event int64) {
	parent = GetEventID()
	event = randInt64()
	SetEventID(event)
	return parent, event
}

func randInt64() int64 {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(fmt.Errorf("could not read random bytes: %v", err))
	}
	// shift to guarantee high bit is 0 and thus
	// int64 version is non-negative
	return int64(binary.BigEndian.Uint64(b[:]) >> 1)
}
