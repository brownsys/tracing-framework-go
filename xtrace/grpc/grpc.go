// Package grpc provides wrappers around certain calls
// in the standard grpc package that automate the propagation
// of X-Trace Task IDs.
package grpc

import (
	"context"
	"fmt"
	"strconv"

	"github.com/brownsys/tracing-framework-go/xtrace/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// MetadataKey is the key used to store an X-Trace
// Task ID in gRPC metadata.
var MetadataKey = "XTRACE_TASKID"

type key struct{}

// Invoke wraps grpc's Invoke function, adding the calling goroutine's
// current Task ID if possible. It first checks client.GetTaskID, and
// if no TaskID is found, it checks ctx for a Task ID stored by this
// package, and if still no Task ID is found, it checks ctx for a Task
// ID received in metadata from an RPC call.
func Invoke(ctx context.Context, method string, args, reply interface{},
	cc *grpc.ClientConn, opts ...grpc.CallOption) error {
	tid := client.GetTaskID()
	if tid == 0 {
		// if ctx.Value(key{}) returns nil,
		// tid will be the zero value - 0
		tid, _ = ctx.Value(key{}).(int64)
	}
	if tid == 0 {
		md, ok := metadata.FromContext(ctx)
		if ok {
			tidstrs := md[MetadataKey]
			if len(tidstrs) > 0 {
				// use ParseUint because Task IDs should never
				// be negative (even though they're stored as
				// int64s)
				t, err := strconv.ParseUint(tidstrs[0], 10, 63)
				if err != nil {
					return fmt.Errorf("parse metadata for key %q (value %q): %v",
						MetadataKey, tidstrs[0], err)
				}
				tid = int64(t)
			}
		}
	}

	if tid != 0 {
		opt := grpc.Header(&metadata.MD{MetadataKey: []string{fmt.Sprint(tid)}})
		// deep copy so we don't modify the caller's copy
		opts = append([]grpc.CallOption(nil), opts...)
		opts = append(opts, opt)
	}

	return grpc.Invoke(ctx, method, args, reply, cc, opts...)
}

// ExtractTaskID attempts to extract a Task ID from
// ctx, and sets it as the calling goroutine's current
// Task ID. It should be called at the beginning of
// any gRPC handler. An error will be returned if a
// Task ID was not found, or if it could not be parsed.
func ExtractTaskID(ctx context.Context) error {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return fmt.Errorf("no metadata found in context")
	}

	tidstrs := md[MetadataKey]
	switch len(tidstrs) {
	case 0:
		return fmt.Errorf("no Task ID found in metadata")
	case 1:
		// use ParseUint because Task IDs should never
		// be negative (even though they're stored as
		// int64s)
		tid, err := strconv.ParseUint(tidstrs[0], 10, 63)
		if err != nil {
			return fmt.Errorf("parse metadata for key %q (value %q): %v",
				MetadataKey, tidstrs[0], err)
		}
		client.SetTaskID(int64(tid))
		return nil
	default:
		return fmt.Errorf("%v > 1 Task IDs found in metadata", len(tidstrs))
	}
}
