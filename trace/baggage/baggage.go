package baggage

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/golang/protobuf/proto"

	lproto "github.com/brownsys/tracing-framework-go/trace/baggage/internal/proto"
)

// ContextKey should be used as the key for baggage
// propagated using a context.Context object (from
// the golang.org/x/net/context (pre-go1.7) or context
// (go1.7) packages).
var ContextKey struct{ private struct{} }

// ByteNamespaces is a set of named namespaces
// in which all values are represented by byte
// slices. Since byte slices are not legal map
// keys in Go, strings with the same contents
// are used instead. These strings are not
// necessarily human-readable, and should be
// treated simply as byte slices which are stored
// temporarily as strings.
type ByteNamespaces map[string]ByteBaggage

// ByteBaggage is a set of named bags
// in which all values are represented by byte
// slices. Since byte slices are not legal map
// keys in Go, strings with the same contents
// are used instead. These strings are not
// necessarily human-readable, and should be
// treated simply as byte slices which are stored
// temporarily as strings.
type ByteBaggage map[string][][]byte

type Marshaler interface {
	MarshalBaggage() ([]byte, error)
}

type Unmarshaler interface {
	UnmarshalBaggage(b []byte) error
}

func Marshal(v interface{}) ([]byte, error) {
	if bv, ok := v.(ByteNamespaces); ok {
		var message lproto.BaggageMessage
		message.Namespace = make([]*lproto.BaggageMessage_NamespaceData, len(bv))

		i := -1
		for k, ns := range bv {
			i++
			var pns lproto.BaggageMessage_NamespaceData
			pns.Key = []byte(k)
			pns.Bag = make([]*lproto.BaggageMessage_BagData, len(ns))
			j := -1
			for k, bag := range ns {
				j++
				var pbag lproto.BaggageMessage_BagData
				pbag.Key = []byte(k)
				pbag.Value = bag
				pns.Bag[j] = &pbag
			}
			message.Namespace[i] = &pns
		}

		bytes, err := proto.Marshal(&message)
		if err != nil {
			return nil, fmt.Errorf("baggage: Marshal: %v", err)
		}
		return bytes, nil
	}

	rv := reflect.ValueOf(v)
	typ := rv.Type()

	marshalSettingsCache.RLock()
	settings, ok := marshalSettingsCache.m[typ]
	marshalSettingsCache.RUnlock()
	if !ok {
		var err error
		settings, err = makeMarshalSettings(typ)
		if err != nil {
			return nil, err
		}

		marshalSettingsCache.Lock()
		marshalSettingsCache.m[typ] = settings
		marshalSettingsCache.Unlock()
	}

	panic("unimplemented")
}

func Unmarshal(data []byte, v interface{}) error {
	bv, ok1 := v.(ByteNamespaces)
	bvp, ok2 := v.(*ByteNamespaces)
	if ok1 || ok2 {
		if ok2 {
			bv = *bvp
		}

		var message lproto.BaggageMessage

		if err := proto.Unmarshal(data, &message); err != nil {
			return fmt.Errorf("baggage: Unmarshal: %v", err)
		}

		for _, ns := range message.GetNamespace() {
			if ns == nil {
				panic("internal error")
			}
			bags := make(ByteBaggage)
			for _, bag := range ns.GetBag() {
				if bag == nil {
					panic("internal error")
				}
				bags[string(bag.GetKey())] = bag.GetValue()
			}
			bv[string(ns.GetKey())] = bags
		}

		return nil
	}

	rv := reflect.ValueOf(v)
	typ := rv.Type()

	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("baggage: Unmarshal non-pointer %v", typ)
	}

	rv = rv.Elem()
	typ = typ.Elem()

	unmarshalSettingsCache.RLock()
	settings, ok := unmarshalSettingsCache.m[typ]
	unmarshalSettingsCache.RUnlock()
	if !ok {
		var err error
		settings, err = makeUnmarshalSettings(typ)
		if err != nil {
			return err
		}

		unmarshalSettingsCache.Lock()
		unmarshalSettingsCache.m[typ] = settings
		unmarshalSettingsCache.Unlock()
	}

	var message lproto.BaggageMessage

	if err := proto.Unmarshal(data, &message); err != nil {
		return fmt.Errorf("baggage: Unmarshal: %v", err)
	}

	for _, ns := range message.GetNamespace() {
		if ns == nil {
			panic("internal error")
		}

		name := reflect.New(settings.namespaceNameTyp)
		if settings.namespaceNameTyp == stringTyp {
			name.Elem().SetString(string(ns.Key))
		} else {
			err := name.Interface().(Unmarshaler).UnmarshalBaggage(ns.Key)
			if err != nil {
				return fmt.Errorf("baggage: Unmarshal: namespace key: %v", err)
			}
		}
		name = name.Elem()

		namespace := reflect.MakeMap(settings.namespaceTyp)
		for _, bag := range ns.GetBag() {
			if bag == nil {
				panic("internal error")
			}

			bk := reflect.New(settings.bagNameTyp)
			if settings.bagNameTyp == stringTyp {
				bk.Elem().SetString(string(bag.Key))
			} else {
				err := bk.Interface().(Unmarshaler).UnmarshalBaggage(bag.Key)
				if err != nil {
					return fmt.Errorf("baggage: Unmarshal: baggage key: %v", err)
				}
			}
			bk = bk.Elem()

			bv := reflect.MakeSlice(reflect.SliceOf(settings.bagElemTyp), len(bag.Value), len(bag.Value))
			for i, v := range bag.Value {
				elem := bv.Index(i)
				if settings.bagElemTyp == byteSliceTyp {
					elem.Set(reflect.ValueOf(v))
				} else {
					err := elem.Addr().Interface().(Unmarshaler).UnmarshalBaggage(v)
					if err != nil {
						return fmt.Errorf("baggage: Unmarshal: baggage element: %v", err)
					}
				}
			}

			namespace.SetMapIndex(bk, bv)
		}

		rv.SetMapIndex(name, namespace)
	}

	return nil
}

var marshalSettingsCache = struct {
	m map[reflect.Type]marshalSettings
	sync.RWMutex
}{m: make(map[reflect.Type]marshalSettings)}

var unmarshalSettingsCache = struct {
	m map[reflect.Type]unmarshalSettings
	sync.RWMutex
}{m: make(map[reflect.Type]unmarshalSettings)}

type marshalSettings struct {
	bagElemTyp       reflect.Type
	bagNameTyp       reflect.Type
	namespaceNameTyp reflect.Type
	namespaceTyp     reflect.Type
}

type unmarshalSettings struct {
	bagElemTyp       reflect.Type
	bagNameTyp       reflect.Type
	namespaceNameTyp reflect.Type
	namespaceTyp     reflect.Type
}

// type converter struct {
// 	bagElemToByteSlice       func(v reflect.Value) []byte
// 	byteSliceToBagElem       func(b []byte) reflect.Value
// 	bagNameToByteSlice       func(v reflect.Value) []byte
// 	byteSliceToBagName       func(b []byte) reflect.Value
// 	namespaceNameToByteSlice func(v reflect.Value) []byte
// 	byteSliceToNamespaceName func(b []byte) reflect.Value
// }

var (
	stringTyp    = reflect.TypeOf("")
	byteSliceTyp = reflect.TypeOf([]byte(nil))

	marshalerTyp   = reflect.TypeOf([]Marshaler(nil)).Elem()
	unmarshalerTyp = reflect.TypeOf([]Unmarshaler(nil)).Elem()
)

func makeMarshalSettings(t reflect.Type) (marshalSettings, error) {
	switch {
	case t.Kind() != reflect.Map:
		fallthrough
	case t.Elem().Kind() != reflect.Map:
		fallthrough
	case t.Elem().Elem().Kind() != reflect.Slice:
		return marshalSettings{}, fmt.Errorf("baggage: Marshal: type must be of the form map[T]map[U][]V")
	}

	var settings marshalSettings

	// namespace name type
	tt := t.Key()
	settings.namespaceTyp = t.Elem()
	settings.namespaceNameTyp = tt
	switch {
	case tt == stringTyp:
	case tt.Implements(marshalerTyp):
	default:
		return marshalSettings{}, fmt.Errorf("baggage: Marshal: type %v is not string and does not implement Marshaler", tt)
	}

	// bage name type
	tt = t.Elem().Key()
	settings.bagNameTyp = tt
	switch {
	case tt == stringTyp:
	case tt.Implements(marshalerTyp):
	default:
		return marshalSettings{}, fmt.Errorf("baggage: Marshal: type %v is not string and does not implement Marshaler", tt)
	}

	// bag element type
	tt = t.Elem().Elem().Elem()
	settings.bagElemTyp = tt
	switch {
	case tt == byteSliceTyp:
	case tt.Implements(marshalerTyp):
	default:
		return marshalSettings{}, fmt.Errorf("baggage: Marshal: type %v is not []byte and does not implement Marshaler", tt)
	}

	return settings, nil
}

func makeUnmarshalSettings(t reflect.Type) (unmarshalSettings, error) {
	switch {
	case t.Kind() != reflect.Map:
		fallthrough
	case t.Elem().Kind() != reflect.Map:
		fallthrough
	case t.Elem().Elem().Kind() != reflect.Slice:
		return unmarshalSettings{}, fmt.Errorf("baggage: Unmarshal: type must be of the form map[T]map[U][]V")
	}

	var settings unmarshalSettings

	// namespace name type
	tt := t.Key()
	settings.namespaceTyp = t.Elem()
	settings.namespaceNameTyp = tt
	switch {
	case tt == stringTyp:
	case reflect.PtrTo(tt).Implements(unmarshalerTyp):
	default:
		return unmarshalSettings{}, fmt.Errorf("baggage: Unmarshal: type %v is not string and does not implement Unmarshaler", tt)
	}

	// bage name type
	tt = t.Elem().Key()
	settings.bagNameTyp = tt
	switch {
	case tt == stringTyp:
	case reflect.PtrTo(tt).Implements(unmarshalerTyp):
	default:
		return unmarshalSettings{}, fmt.Errorf("baggage: Unmarshal: type %v is not string and does not implement Unmarshaler", tt)
	}

	// bag element type
	tt = t.Elem().Elem().Elem()
	settings.bagElemTyp = tt
	switch {
	case tt == byteSliceTyp:
	case reflect.PtrTo(tt).Implements(unmarshalerTyp):
	default:
		return unmarshalSettings{}, fmt.Errorf("baggage: Unmarshal: type %v is not []byte and does not implement Unmarshaler", tt)
	}

	return settings, nil
}
