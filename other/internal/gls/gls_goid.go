// +build goid

package gls

import (
	"runtime"
	"sync"
)

type bucket struct {
	locals map[int64]*local
	sync.RWMutex
}

type local struct {
	m map[interface{}]interface{}
	sync.RWMutex
}

var buckets = make([]bucket, 8*runtime.NumCPU())

func init() {
	for i := range buckets {
		buckets[i].locals = make(map[int64]*local)
	}
}

func getLocal() *local {
	goid := runtime.Goid()
	bkt := &buckets[int(goid%int64(len(buckets)))]
	bkt.RLock()
	l, ok := bkt.locals[goid]
	bkt.RUnlock()
	if !ok {
		panic("gls: no goroutine-local storage; goroutine was not spawned with Go")
	}
	return l
}

func _go(f func()) {
	go base(f)
}

func base(f func()) {
	goid := runtime.Goid()
	bkt := &buckets[int(goid%int64(len(buckets)))]

	l := &local{m: make(map[interface{}]interface{})}
	bkt.Lock()
	bkt.locals[goid] = l
	bkt.Unlock()

	defer func() {
		bkt.Lock()
		delete(bkt.locals, goid)
		bkt.Unlock()
	}()

	f()
}

func put(k, v interface{}) {
	l := getLocal()
	l.Lock()
	// defer in case k is an invalid
	// map key and the next line panics
	defer l.Unlock()
	l.m[k] = v
}

func get(k interface{}) (interface{}, bool) {
	l := getLocal()
	l.RLock()
	// defer in case k is an invalid
	// map key and the next line panics
	defer l.RUnlock()
	v, ok := l.m[k]
	return v, ok
}

func del(k interface{}) {
	l := getLocal()
	l.Lock()
	// defer in case k is an invalid
	// map key and the next line panics
	defer l.Unlock()
	delete(l.m, k)
}
