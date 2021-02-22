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
		binary.BigEndian.PutUint16(arr[1:], uint16(v.Value))
		return arr
	}
	arr := make([]byte, 5)
	arr[0] = 0xf0
	binary.BigEndian.PutUint32(arr[1:], uint32(v.Value))
	return arr
}

func (p *Property) Encode() []byte {
	var buf bytes.Buffer
	if p.Name+0x80 < 0xa0 {
		val := uint8(0x80 + p.Name)
		binary.Write(&buf, binary.BigEndian, val)
	} else if p.Name < 0x100 {
		val := uint8(p.Name)
		buf.Write([]byte("\xa0"))
		binary.Write(&buf, binary.BigEndian, val)
	} else {
		val := uint16(p.Name)
		buf.Write([]byte("\xc0"))
		binary.Write(&buf, binary.BigEndian, val)
	}
	v := Number{p.Value}
	buf.Write(v.Encode())
	return buf.Bytes()
}

func (n *Node) Encode() []byte {
	var buf bytes.Buffer
	buf.Write([]byte("\x00"))
	name_index := Number{n.Name}
	buf.Write(name_index.Encode())
	np := Number{len(n.Properties)}
	buf.Write(np.Encode())
	nc := Number{len(n.Children)}
	buf.Write(nc.Encode())
	// props
	for _, p := range n.Properties {
		buf.Write(p.Encode())
	}
	// child nodes
	for _, c := range n.Children {
		buf.Write(c.Encode())
	}
	return buf.Bytes()
}

func (d *Document) Encode() []byte {
	var buf bytes.Buffer
	buf.Write([]byte("IBX1"))
	// strings
	nstrings := Number{len(d.Strings)}
	buf.Write(nstrings.Encode())
	for _, s := range d.Strings {
		n := Number{len(s)}
		buf.Write(n.Encode())
		buf.Write([]byte(s))
		buf.Write([]byte("\x00"))
	}
	// typed values
	nvals := Number{len(d.TypedValues)}
	buf.Write(nvals.Encode())
	for _, tv := range d.TypedValues {
		buf.Write(tv.Encode())
	}
	// encoding flag
	buf.Write([]byte("\x01"))
	// node structure
	buf.Write(d.Element.Encode())
	return buf.Bytes()
}
