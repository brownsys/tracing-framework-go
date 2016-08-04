package local

import (
	"fmt"
	"runtime"
)

type local []interface{}

var callbacks []Callbacks
var defaults []interface{}

// Token represents a handle on a particular
// local variable; it is used to distinguish
// a local variable from others so that different
// packages can use goroutine-local storage
// simultaneously without interfering with
// each other.
type Token token

// make sure that Token is an opaque type
// so that the caller cannot manipulate it,
// and we can change it later and not break
// external code
type token int

// Callbacks is a collection of functions that
// should be used for modifying local values.
type Callbacks struct {
	// LocalForSpawn takes the current goroutine's
	// local and produces a local that should be
	// set in the spawned goroutine. If LocalForSpawn
	// is nil, the current goroutine's local is used
	// as the initial value in the spawned goroutine.
	LocalForSpawn func(local interface{}) interface{}
}

// Register registers a new local variable whose
// initial value and callbacks are set from the
// arguments. The returned Token can be used to
// identify the variable in future calls.
//
// Register should ONLY be called during initialization,
// and in the main goroutine.
func Register(initial interface{}, c Callbacks) Token {
	defaults = append(defaults, initial)
	callbacks = append(callbacks, c)
	return Token(len(callbacks) - 1)
}

// GetSpawnCallback returns a function which should
// be called in the spawned goroutine in order to
// set the local variables appropriately.
//
// NOTE: This should only be called by code generated
// with the associated rewrite tool.
func GetSpawnCallback() func() {
	l := getLocal()
	newl := make(local, len(l))
	for i, c := range callbacks {
		f := c.LocalForSpawn
		if f == nil {
			newl[i] = l[i]
		} else {
			newl[i] = f(l[i])
		}
	}

	return func() {
		runtime.SetLocal(newl)
	}
}

// GetLocal returns the local variable associated
// with the given Token.
func GetLocal(t Token) interface{} {
	l := getLocal()
	return l[t]
}

// SetLocal sets the local variable associated
// with the given Token.
func SetLocal(t Token, l interface{}) {
	ll := getLocal()
	ll[t] = l
}

// getLocal retreives this goroutine's
// locals slice, initializing it to
// a copy of defaults if none exists
func getLocal() local {
	l, ok := runtime.GetLocal().(local)
	if !ok {
		l = make(local, len(defaults))
		copy(l, defaults)
		runtime.SetLocal(l)
	}
	if len(l) != len(defaults) {
		panic(fmt.Errorf("unexpected number of locals: got %v; want %v", len(l), len(defaults)))
	}
	return l
}
