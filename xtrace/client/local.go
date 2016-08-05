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

func newEvent() (parent, event int64) {
	parent = local.GetLocal(token).(int64)
	event = randInt64()
	local.SetLocal(token, event)
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
