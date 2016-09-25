// +build !goid

package gls

import (
	"fmt"
	"runtime"
	"unsafe"
)

func _go(f func()) {
	go base(f)
}

type shim func([]shim)

func base(f func()) {
	m := make(map[interface{}]interface{})
	ptr := uintptr(unsafe.Pointer(&m))

	var s []shim

	switch runtime.GOARCH {
	case "386":
		s = make([]shim, 4)
		s[0] = shims[int((ptr>>8)&0xFF)]
		s[1] = shims[int((ptr>>16)&0xFF)]
		s[2] = shims[int((ptr>>24)&0xFF)]
	case "amd64":
		s = make([]shim, 8)
		s[0] = shims[int((ptr>>8)&0xFF)]
		s[1] = shims[int((ptr>>16)&0xFF)]
		s[2] = shims[int((ptr>>24)&0xFF)]
		s[3] = shims[int((ptr>>32)&0xFF)]
		s[4] = shims[int((ptr>>40)&0xFF)]
		s[5] = shims[int((ptr>>48)&0xFF)]
		s[6] = shims[int((ptr>>56)&0xFF)]
	default:
		panic(fmt.Errorf("gls: unsupported GOARCH %v", runtime.GOARCH))
	}
	s[len(s)-1] = func(_ []shim) { f() }

	shims[int(ptr&0xFF)](s)

	// use m so the compiler can't optimize
	// it away (since we're using an unsafe pointer)
	m[1] = 1
}

var basePtr uintptr

func init() {
	skip := 3
	switch runtime.GOARCH {
	case "386":
		skip += 4
	case "amd64":
		skip += 8
	default:
		panic(fmt.Errorf("gls: unsupported GOARCH %v", runtime.GOARCH))
	}
	f := func() {
		pcs := make([]uintptr, 16)
		runtime.Callers(skip, pcs)
		basePtr = pcs[0]
	}
	base(f)
}

var pcToUintptr = make(map[uintptr]uintptr)

var shimPCOffset uintptr

// getPtr computes the current goroutine's
// goroutine-local pointer
func getPtr() uintptr {
	var pcs [16]uintptr
	// we know we can at least skip this
	// call and the call to getPtr
	i := runtime.Callers(2, pcs[:])

	if i < 2 || pcs[i-2] != basePtr {
		panic("gls: no local storage associated with this goroutine")
	}
	// we don't care about runtime.goexit or base,
	// and we want i to be the index of the last
	// element, not the length
	i -= 3

	var ptr uintptr
	switch runtime.GOARCH {
	case "386":
		ptr |= pcToUintptr[pcs[i]]
		ptr |= pcToUintptr[pcs[i-1]] << 8
		ptr |= pcToUintptr[pcs[i-2]] << 16
		ptr |= pcToUintptr[pcs[i-3]] << 24
	case "amd64":
		ptr |= pcToUintptr[pcs[i]]
		ptr |= pcToUintptr[pcs[i-1]] << 8
		ptr |= pcToUintptr[pcs[i-2]] << 16
		ptr |= pcToUintptr[pcs[i-3]] << 24
		ptr |= pcToUintptr[pcs[i-4]] << 32
		ptr |= pcToUintptr[pcs[i-5]] << 40
		ptr |= pcToUintptr[pcs[i-6]] << 48
		ptr |= pcToUintptr[pcs[i-7]] << 56
	default:
		panic(fmt.Errorf("gls: unsupported GOARCH %v", runtime.GOARCH))
	}

	return ptr
}

func put(k, v interface{}) {
	ptr := getPtr()
	m := *(*map[interface{}]interface{})(unsafe.Pointer(ptr))
	m[k] = v
}

func get(k interface{}) (interface{}, bool) {
	ptr := getPtr()
	m := *(*map[interface{}]interface{})(unsafe.Pointer(ptr))
	v, ok := m[k]
	return v, ok
}

func del(k interface{}) {
	ptr := getPtr()
	m := *(*map[interface{}]interface{})(unsafe.Pointer(ptr))
	delete(m, k)
}
