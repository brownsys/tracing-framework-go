package local

import (
	"sync"
	"testing"
)

func BenchmarkGo(b *testing.B) {
	var wg sync.WaitGroup
	wg.Add(b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go func() { wg.Done() }()
	}
	wg.Wait()
}

func BenchmarkGetSpawnCallback(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetSpawnCallback()
	}
}
