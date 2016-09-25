package runtime_test

import (
	"math/rand"
	"reflect"
	"runtime"
	"testing"
)

func TestLocal(t *testing.T) {
	for i := 0; i < 1024; i++ {
		want := rand.Uint32()
		runtime.SetLocal(want)
		got, ok := runtime.GetLocal().(uint32)
		if !ok {
			t.Fatalf("unexpected local type: got %v; want %v", reflect.TypeOf(got), reflect.TypeOf(want))
		}
		if got != want {
			t.Fatalf("unexpected local value: got %v; want %v", got, want)
		}
	}
}
