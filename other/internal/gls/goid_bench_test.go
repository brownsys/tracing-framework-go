// +build goid

package gls

import (
	"runtime"
	"sync"
	"testing"
)

func BenchmarkGoid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runtime.Goid()
	}
}

func BenchmarkSpawnGoid(b *testing.B) {
	var wg sync.WaitGroup
	wg.Add(b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Go(func() { wg.Done() })
	}
	wg.Wait()
}

func BenchmarkPutGoid(b *testing.B) {
	nroutines := runtime.NumCPU() * 4
	var wg sync.WaitGroup
	wg.Add(nroutines)
	b.ResetTimer()
	for i := 0; i < nroutines; i++ {
		Go(func() {
			for i := 0; i < b.N; i++ {
				Put(1, 1)
			}
			wg.Done()
		})
	}
	wg.Wait()
}

func BenchmarkGetGoid(b *testing.B) {
	nroutines := runtime.NumCPU() * 4
	var wg sync.WaitGroup
	wg.Add(nroutines)
	b.ResetTimer()
	for i := 0; i < nroutines; i++ {
		Go(func() {
			for i := 0; i < b.N; i++ {
				Get(1)
			}
			wg.Done()
		})
	}
	wg.Wait()
}

func BenchmarkDeleteGoid(b *testing.B) {
	nroutines := runtime.NumCPU() * 4
	var wg sync.WaitGroup
	wg.Add(nroutines)
	b.ResetTimer()
	for i := 0; i < nroutines; i++ {
		Go(func() {
			for i := 0; i < b.N; i++ {
				Delete(1)
			}
			wg.Done()
		})
	}
	wg.Wait()
}
