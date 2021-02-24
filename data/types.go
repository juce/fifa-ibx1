package data

import (
	"fmt"
)

type Int8 struct {
	Value int8
}

type Int16 struct {
	Value int16
}

type Int32 struct {
	Value int32
}

type UInt8 struct {
	Value uint8
}

type UInt16 struct {
	Value uint16
}

type UInt32 struct {
	Value uint32
}

type String struct {
	Value int
}

type Bool struct {
	Value bool
}

type Float struct {
	Value float32
}

func (v String) TypeId() int {
	if v.Value < 0x10 {
		return 0xc0 + v.Value
	}
	if v.Value < 0x100 {
		return 0xd0
	}
	return 0xe0
}

func (v Int8) TypeId() int {
	if v.Value >= 0 && v.Value < 0x10 {
		return int(v.Value)
	}
	return 0x10
}

func (v UInt8) TypeId() int {
	if v.Value >= 0 && v.Value < 0x10 {
		return int(v.Value)
	}
	return 0x10
}

func (v Int16) TypeId() int {
	return 0x20
}

func (v UInt16) TypeId() int {
	return 0x20
}

func (v Int32) TypeId() int {
	return 0x30
}

func (v UInt32) TypeId() int {
	return 0x30
}

func (v Bool) TypeId() int {
	if v.Value {
		return 0x41
	}
	return 0x40
}

func (v Float) TypeId() int {
	return 0xb0
}

func (v Int8) String() string {
	return fmt.Sprintf("{0x%02x %s %d}", v.TypeId(), "int8", v.Value)
}

func (v Int16) String() string {
	return fmt.Sprintf("{0x%02x %s %d}", v.TypeId(), "int16", v.Value)
}

func (v Int32) String() string {
	return fmt.Sprintf("{0x%02x %s %d}", v.TypeId(), "int32", v.Value)
}

func (v UInt8) String() string {
	return fmt.Sprintf("{0x%02x %s %d}", v.TypeId(), "uint8", v.Value)
}

func (v UInt16) String() string {
	return fmt.Sprintf("{0x%02x %s %d}", v.TypeId(), "uint16", v.Value)
}

func (v UInt32) String() string {
	return fmt.Sprintf("{0x%02x %s %d}", v.TypeId(), "uint32", v.Value)
}

func (v String) String() string {
	return fmt.Sprintf("{0x%02x %s %d}", v.TypeId(), "string", v.Value)
}

func (v Bool) String() string {
	return fmt.Sprintf("{0x%02x %s %v}", v.TypeId(), "bool", v.Value)
}

func (v Float) String() string {
	return fmt.Sprintf("{0x%02x %s %f}", v.TypeId(), "float", v.Value)
}

func (v String) Deref(strings []string) string {
	return fmt.Sprintf("{0x%02x %s '%s'}", v.TypeId(), "string", strings[v.Value])
}
