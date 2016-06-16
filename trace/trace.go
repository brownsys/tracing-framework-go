package trace

import (
	"fmt"
	"local/research/baggage"
	"reflect"
	"runtime"

	"golang.org/x/net/context"

	"github.com/brownsys/tracing-framework-go/trace/internal/instrument"
)

func GetArgTypes(f interface{}) (args []reflect.Type, variadic, ok bool) {
	return GetArgTypesName(interfaceToName(f, "GetArgTypes"))
}

func GetArgTypesName(fname string) (args []reflect.Type, variadic, ok bool) {
	typ, ok := instrument.GetTypeName(fname)
	if !ok {
		return nil, false, false
	}
	args = make([]reflect.Type, typ.NumIn())
	for i := 0; i < typ.NumIn(); i++ {
		args[i] = typ.In(i)
	}
	return args, typ.IsVariadic(), true
}

func getArgsFromType(typ reflect.Type) ([]reflect.Type, bool) {
	args := make([]reflect.Type, typ.NumIn())
	for i := 0; i < typ.NumIn(); i++ {
		args[i] = typ.In(i)
	}
	return args, typ.IsVariadic()
}

func Instrument(f interface{}, callback func(bag interface{}, args []reflect.Value)) {
	InstrumentName(interfaceToName(f, "Instrument"), callback)
}

func InstrumentName(fname string, callback func(bag interface{}, args []reflect.Value)) {
	typ, _ := instrument.GetTypeName(fname)
	f := func(args []reflect.Value) []reflect.Value {
		callback(args[0].Interface().(context.Context).Value(baggage.ContextKey), args[1:])
		return nil
	}
	instrument.InstrumentName(fname, reflect.MakeFunc(typ, f).Interface())
}

func Uninstrument(f interface{}) {
	instrument.Uninstrument(f)
}

func UninstrumentName(fname string) {
	instrument.UninstrumentName(fname)
}

// f is the function whose name should be retreived,
// and fname is the name of the top-level function
// that is calling interfaceToName (used in panic
// messages)
func interfaceToName(f interface{}, fname string) string {
	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Func {
		panic(fmt.Errorf("trace: %v with non-func %v", fname, v.Type()))
	}
	return runtime.FuncForPC(v.Pointer()).Name()
}
