package layout

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestFieldsUnchanged(t *testing.T) {
	want := map[string]reflect.Type{
		"Active":   reflect.TypeFor[bool](),
		"ID":       reflect.TypeFor[int64](),
		"Kind":     reflect.TypeFor[uint8](),
		"Score":    reflect.TypeFor[float64](),
		"Retries":  reflect.TypeFor[int16](),
		"Name":     reflect.TypeFor[string](),
		"Priority": reflect.TypeFor[int32](),
		"Deleted":  reflect.TypeFor[bool](),
	}
	typ := reflect.TypeFor[Event]()
	if typ.NumField() != len(want) {
		t.Fatalf("Event has %d fields, want %d", typ.NumField(), len(want))
	}
	for i := range typ.NumField() {
		f := typ.Field(i)
		if want[f.Name] != f.Type {
			t.Errorf("field %s has type %v, want %v", f.Name, f.Type, want[f.Name])
		}
	}
}

func TestSize(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("sizes below assume a 64-bit platform")
	}
	if got := unsafe.Sizeof(Event{}); got > 48 {
		t.Errorf("unsafe.Sizeof(Event{}) = %d bytes, want 48 (it can be done)", got)
	}
}
