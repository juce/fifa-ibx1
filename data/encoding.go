package data

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

func (n Number) Encode() []byte {
	var arr []byte
	if n.Value < 0x40 {
		arr = make([]byte, 1)
		arr[0] = byte(n.Value)
		return arr
	}
	if n.Value < 0x100 {
		arr = make([]byte, 2)
		arr[0] = 0x40
		arr[1] = byte(n.Value)
		return arr
	}
	if n.Value < 0x10000 {
		arr = make([]byte, 3)
		arr[0] = 0x80
		binary.BigEndian.PutUint16(arr[1:], uint16(n.Value))
		return arr
	}
	if n.Value < 0x100000000 {
		arr = make([]byte, 5)
		arr[0] = 0xc0
		binary.BigEndian.PutUint32(arr[1:], uint32(n.Value))
		return arr
	}
	arr = make([]byte, 9)
	arr[0] = 0xf0
	binary.BigEndian.PutUint64(arr[1:], uint64(n.Value))
	return arr
}

func (n Number) Hex() string {
	bs := n.Encode()
	parts := make([]string, len(bs))
	for i, b := range bs {
		parts[i] = fmt.Sprintf("%02x", b)
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

func (v Int8) Encode() []byte {
	if v.Value >= 0 && v.Value < 0x10 {
		arr := make([]byte, 1)
		arr[0] = byte(v.Value)
		return arr
	}
	arr := make([]byte, 2)
	arr[0] = 0x10
	arr[1] = byte(v.Value)
	return arr
}

func (v Int16) Encode() []byte {
	arr := make([]byte, 3)
	arr[0] = 0x20
	binary.LittleEndian.PutUint16(arr[1:], uint16(v.Value))
	return arr
}

func (v Int32) Encode() []byte {
	arr := make([]byte, 5)
	arr[0] = 0x30
	binary.LittleEndian.PutUint32(arr[1:], uint32(v.Value))
	return arr
}

func (v UInt8) Encode() []byte {
	if v.Value >= 0 && v.Value < 0x10 {
		arr := make([]byte, 1)
		arr[0] = byte(v.Value)
		return arr
	}
	arr := make([]byte, 2)
	arr[0] = 0x10
	arr[1] = byte(v.Value)
	return arr
}

func (v UInt16) Encode() []byte {
	arr := make([]byte, 3)
	arr[0] = 0x20
	binary.LittleEndian.PutUint16(arr[1:], v.Value)
	return arr
}

func (v UInt32) Encode() []byte {
	arr := make([]byte, 5)
	arr[0] = 0x30
	binary.LittleEndian.PutUint32(arr[1:], v.Value)
	return arr
}

func (v Bool) Encode() []byte {
	arr := make([]byte, 1)
	if v.Value == false {
		arr[0] = 0x40
	} else {
		arr[0] = 0x41
	}
	return arr
}

func (v Float) Encode() []byte {
	var buf bytes.Buffer
	buf.Write([]byte{0xb0})
	binary.Write(&buf, binary.LittleEndian, v.Value)
	return buf.Bytes()
}

func (v String) Encode() []byte {
	if v.Value < 0x10 {
		return []byte{byte(0xc0 + v.Value)}
	}
	if v.Value < 0x100 {
		return []byte{0xd0, byte(v.Value)}
	}
	if v.Value < 0x10000 {
		arr := make([]byte, 3)
		arr[0] = 0xe0
		binary.LittleEndian.PutUint16(arr[1:], uint16(v.Value))
		return arr
	}
	arr := make([]byte, 5)
	arr[0] = 0xf0
	binary.LittleEndian.PutUint32(arr[1:], uint32(v.Value))
	return arr
}
