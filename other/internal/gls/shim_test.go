// +build !goid

package gls

import (
	"math/rand"
	"runtime"
	"sync"
	"testing"
)

func TestPointerResolution(t *testing.T) {
	vals := rand.Perm(1024)
	var wg sync.WaitGroup
	for i := 0; i < 4*runtime.NumCPU(); i++ {
		wg.Add(1)
		f := func() {
			for i, v := range vals {
				Put(i, v)
			}
			for i, v := range vals {
				got, ok := Get(i)

				if !ok || got != v {
					t.Errorf("unexpected result: got %v:%v; want true:%v", ok, got, v)
				}
			}
			wg.Done()
		}
		Go(f)
	}
	wg.Wait()
}
