package client

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/brownsys/tracing-framework-go/local"
)

var token local.Token

func init() {
	token = local.Register(int64(0), local.Callbacks{
		func(l interface{}) interface{} { return l },
	})
}

// SetEventID sets the current goroutine's X-Trace Task ID.
// This should be used when propagating Task IDs over RPC
// calls or other channels.
//
// WARNING: This will overwrite any previous Task ID,
// so call with caution.
func SetTaskID(taskID int64) {
	local.SetLocal(token, taskID)
}

// GetTaskID gets the current goroutine's X-Trace Task ID.
// Note that if one has not been set yet, GetTaskID will
// return 0. This should be used when propagating Task IDs
// over RPC calls or other channels.
func GetTaskID() (taskID int64) {
	return local.GetLocal(token).(int64)
}

func newEvent() (parent, event int64) {
	parent = GetTaskID()
	event = randInt64()
	SetTaskID(event)
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
