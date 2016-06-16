package baggage

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"

	lproto "github.com/brownsys/tracing-framework-go/trace/baggage/internal/proto"
)

func TestUnmarshalBasic(t *testing.T) {
	testCaseUnmarshal(t, nil, ByteNamespaces{})
	testCaseUnmarshal(t, []testNamespace{{}}, ByteNamespaces{"": ByteBaggage{}})
	testCaseUnmarshal(t, []testNamespace{{key: []byte("foo"),
		vals: []testBag{{key: []byte("bar"), vals: [][]byte{[]byte("baz")}}}}},
		ByteNamespaces{"foo": ByteBaggage{"bar": [][]byte{[]byte("baz")}}})
}

type unmarshalableUint32 uint32

func (u *unmarshalableUint32) UnmarshalBaggage(buf []byte) error {
	if len(buf) != 4 {
		return fmt.Errorf("invalid buffer length: %v", len(buf))
	}
	*u = unmarshalableUint32(binary.BigEndian.Uint32(buf))
	return nil
}

func TestUnmarshalInterface(t *testing.T) {
	type tt1 map[string][]unmarshalableUint32
	type t1 map[string]tt1
	testCaseUnmarshal(t, nil, t1{})
	testCaseUnmarshal(t, []testNamespace{{}}, t1{"": tt1{}})
	testCaseUnmarshal(t, []testNamespace{{key: []byte("foo"),
		vals: []testBag{{key: []byte("bar"), vals: [][]byte{[]byte{1, 2, 3, 4}}}}}},
		t1{"foo": tt1{"bar": []unmarshalableUint32{0x1020304}}})

	type tt2 map[unmarshalableUint32][][]byte
	type t2 map[string]tt2
	testCaseUnmarshal(t, nil, t2{})
	testCaseUnmarshal(t, []testNamespace{{}}, t2{"": tt2{}})
	testCaseUnmarshal(t, []testNamespace{{key: []byte("foo"),
		vals: []testBag{{key: []byte{0, 1, 2, 3}, vals: [][]byte{[]byte("bar")}},
			{key: []byte{0, 1, 2, 3}, vals: [][]byte{[]byte("baz")}}}}},
		t2{"foo": tt2{0x10203: [][]byte{[]byte("baz")}}})

	type tt3 map[string][][]byte
	type t3 map[unmarshalableUint32]tt3
	testCaseUnmarshal(t, nil, t3{})
	testCaseUnmarshal(t, []testNamespace{{key: []byte{0, 1, 2, 3}, vals: []testBag{}},
		{key: []byte{0, 1, 2, 3}, vals: []testBag{{key: []byte("foo"), vals: [][]byte{[]byte("bar")}}}}},
		t3{0x10203: tt3{"foo": [][]byte{[]byte("bar")}}})
}

// expect must be a valid type (map[T]map[U][]V)
func testCaseUnmarshal(t *testing.T, ns []testNamespace, expect interface{}) {
	msg := testNamespaceToProto(ns)
	buf, err := proto.Marshal(&msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	typ := reflect.TypeOf(expect)
	got := reflect.New(typ)
	got.Elem().Set(reflect.MakeMap(typ))
	err = Unmarshal(buf, got.Interface())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if !reflect.DeepEqual(got.Elem().Interface(), expect) {
		t.Errorf("unexpected result: got %#v; want %#v", got.Elem().Interface(), expect)
	}
}

type testNamespace struct {
	key  []byte
	vals []testBag
}

type testBag struct {
	key  []byte
	vals [][]byte
}

func testNamespaceToProto(t []testNamespace) lproto.BaggageMessage {
	var msg lproto.BaggageMessage
	for _, ns := range t {
		var pns lproto.BaggageMessage_NamespaceData
		// always allocate even if ns.key == nil
		pns.Key = append([]byte{}, ns.key...)
		for _, bag := range ns.vals {
			var pbag lproto.BaggageMessage_BagData
			// always allocate even if bag.key == nil
			pbag.Key = append([]byte{}, bag.key...)
			pbag.Value = bag.vals
			pns.Bag = append(pns.Bag, &pbag)
		}
		msg.Namespace = append(msg.Namespace, &pns)
	}
	return msg
}
