package gls

import (
	"runtime"
	"sync"
	"testing"
)

func BenchmarkCallers(b *testing.B) {
	pc := make([]uintptr, 64)
	b.ResetTimer()
	benchmarkCallers(1, b.N, pc)
}

func BenchmarkCallers2(b *testing.B) {
	pc := make([]uintptr, 64)
	b.ResetTimer()
	benchmarkCallers(2, b.N, pc)
}

func BenchmarkCallers4(b *testing.B) {
	pc := make([]uintptr, 64)
	b.ResetTimer()
	benchmarkCallers(4, b.N, pc)
}

func BenchmarkCallers8(b *testing.B) {
	pc := make([]uintptr, 64)
	b.ResetTimer()
	benchmarkCallers(8, b.N, pc)
}

func benchmarkCallers(depth, n int, pc []uintptr) {
	if depth > 1 {
		benchmarkCallers(depth-1, n, pc)
		return
	}
	for i := 0; i < n; i++ {
		runtime.Callers(0, pc)
	}
}

func BenchmarkLockMapAccessReadOnly(b *testing.B) {
	m := make(map[int]int)
	m[1] = 1
	var mtx sync.RWMutex

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mtx.RLock()
		_ = m[1]
		mtx.RUnlock()
	}
}

func BenchmarkLockMapAccessReadWrite(b *testing.B) {
	m := make(map[int]int)
	m[1] = 1
	var mtx sync.RWMutex

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mtx.Lock()
		_ = m[1]
		mtx.Unlock()
	}
}

func BenchmarkSpawnNormal(b *testing.B) {
	var wg sync.WaitGroup
	wg.Add(b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go func() { wg.Done() }()
	}
	wg.Wait()
}

// func BenchmarkLockMapAccessReadOnlyContention(b *testing.B) {
// 	m := make(map[int]int)
// 	m[1] = 1
// 	var mtx sync.RWMutex

// 	var wg sync.WaitGroup

// 	f := func() {
// 		for i := 0; i < b.N; i++ {
// 			mtx.RLock()
// 			_ = m[1]
// 			mtx.RUnlock()
// 		}
// 		wg.Done()
// 	}

// 	b.ResetTimer()
// 	for i := 0; i < runtime.NumCPU(); i++ {
// 		wg.Add(1)
// 		// Use Go so there's an even amount of
// 		// goroutine creation overhead
// 		Go(f, 0)
// 	}
// 	wg.Wait()
// }
