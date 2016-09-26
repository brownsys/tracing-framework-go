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

var (
	// EventIDMetadataKey is the key used to store an
	// X-Trace Event ID in gRPC metadata.
	EventIDMetadataKey = "xtrace_event_id"
	// TaskIDMetadataKey is the key used to store an
	// X-Trace Task ID in gRPC metadata.
	TaskIDMetadataKey = "xtrace_task_id"
)

// Invoke wraps grpc's Invoke function, adding the calling goroutine's
// current Task and Event IDs if possible.
func Invoke(ctx context.Context, method string, args, reply interface{},
	cc *grpc.ClientConn, opts ...grpc.CallOption) error {
	eid := client.GetEventID()
	tid := client.GetTaskID()

	var pairs []string
	if eid != 0 {
		pairs = append(pairs, EventIDMetadataKey, fmt.Sprint(eid))
	}
	if tid != 0 {
		pairs = append(pairs, TaskIDMetadataKey, fmt.Sprint(tid))
	}
	if eid != 0 || tid != 0 {
		md := metadata.Pairs(pairs...)
		ctx = metadata.NewContext(ctx, md)
	}
	return grpc.Invoke(ctx, method, args, reply, cc, opts...)
}

// ExtractIDs attempts to extract a Task ID and an
// Event ID from ctx, and sets it as the calling
// goroutine's current IDs. It should be called at
// the beginning of any gRPC handler. An error will
// be returned if either ID was not found, or if it
// could not be parsed. Note that if one ID was found
// and parsed successfully, but the other was not,
// neither will be set.
func ExtractIDs(ctx context.Context) error {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return fmt.Errorf("no metadata found in context")
	}

	eid, err := getIDFromMetadata(md, EventIDMetadataKey, "Event ID")
	if err != nil {
		return err
	}
	tid, err := getIDFromMetadata(md, TaskIDMetadataKey, "Task ID")
	if err != nil {
		return err
	}

	client.SetEventID(eid)
	client.SetTaskID(tid)
	return nil
}

// get the given ID from md; name should be "Task ID" or "Event ID",
// and getIDFromMetadata will produce the proper error messages that
// can be returned directly without wrapping.
func getIDFromMetadata(md metadata.MD, key, name string) (int64, error) {
	m := make(map[string]string)
	for k, v := range md {
		for _, v := range v {
			kk, vv, err := metadata.DecodeKeyValue(k, v)
			if err != nil {
				return 0, fmt.Errorf("malformed metadata: %v", err)
			}
			m[kk] = vv
		}
	}
	str, ok := m[key]
	if !ok {
		return 0, fmt.Errorf("no %v found in metadata", name)
	}
	// use ParseUint because IDs should never be
	// negative (even though they're stored as
	// int64s)
	id, err := strconv.ParseUint(str, 10, 63)
	if err != nil {
		return 0, fmt.Errorf("parse metadata for key %q (value %q): %v", key, str, err)
	}
	return int64(id), nil
}
