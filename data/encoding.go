package data

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

type Options struct {
	Hex8  bool
	Hex16 bool
	Hex32 bool
}

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

func ReadNumber(reader *bufio.Reader) (*Number, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	if b < 0x40 {
		return &Number{int(b)}, nil
	}
	if b == 0x40 {
		v, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		return &Number{int(v)}, nil
	} else if b == 0x80 {
		var bs = make([]byte, 2)
		_, err := io.ReadFull(reader, bs)
		if err != nil {
			return nil, err
		}
		v := binary.BigEndian.Uint16(bs)
		return &Number{int(v)}, nil
	}
	return nil, fmt.Errorf("unknown number encoding")
}

func ReadTypedValue(reader *bufio.Reader) (TypedValue, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	if b < 0x10 {
		return Int8{int8(b)}, nil
	}
	if b == 0x10 {
		v, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		return Int8{int8(v)}, nil
	}
	if b == 0x20 {
		bs := make([]byte, 2)
		_, err := io.ReadFull(reader, bs)
		if err != nil {
			return nil, err
		}
		v := binary.LittleEndian.Uint16(bs)
		return Int16{int16(v)}, nil
	}
	if b == 0x30 {
		bs := make([]byte, 4)
		_, err := io.ReadFull(reader, bs)
		if err != nil {
			return nil, err
		}
		v := binary.LittleEndian.Uint32(bs)
		return Int32{int32(v)}, nil
	}
	if b == 0x40 {
		return Bool{false}, nil
	}
	if b == 0x41 {
		return Bool{true}, nil
	}
	if b == 0xb0 {
		bs := make([]byte, 4)
		_, err := io.ReadFull(reader, bs)
		if err != nil {
			return nil, err
		}
		var v float32
		buf := bytes.NewReader(bs)
		err = binary.Read(buf, binary.LittleEndian, &v)
		if err != nil {
			return nil, err
		}
		return Float{v}, nil
	}
	if b >= 0xc0 && b < 0xd0 {
		v := int(b) - 0xc0
		return String{v}, nil
	}
	if b == 0xd0 {
		v, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		return String{int(v)}, nil
	}
	if b == 0xe0 {
		bs := make([]byte, 2)
		_, err := io.ReadFull(reader, bs)
		if err != nil {
			return nil, err
		}
		var v uint16
		buf := bytes.NewReader(bs)
		err = binary.Read(buf, binary.BigEndian, &v)
		if err != nil {
			return nil, err
		}
		return String{int(v)}, nil
	}

	return nil, nil //fmt.Errorf("unknown typed-value encoding")
}

func ReadProperty(reader *bufio.Reader) (*Property, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	if b < 0xa0 {
		nameIndex := int(b) - 0x80
		v, err := ReadNumber(reader)
		if err != nil {
			return nil, err
		}
		return &Property{Name: nameIndex, Value: v.Value}, nil
	}
	if b == 0xa0 {
		v, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		nameIndex := int(v)
		val, err := ReadNumber(reader)
		if err != nil {
			return nil, err
		}
		return &Property{Name: nameIndex, Value: val.Value}, nil
	}
	if b == 0xc0 {
		bs := make([]byte, 2)
		_, err := io.ReadFull(reader, bs)
		if err != nil {
			return nil, err
		}
		var v uint16
		buf := bytes.NewReader(bs)
		err = binary.Read(buf, binary.BigEndian, &v)
		if err != nil {
			return nil, err
		}
		nameIndex := int(v)
		val, err := ReadNumber(reader)
		if err != nil {
			return nil, err
		}
		return &Property{Name: nameIndex, Value: val.Value}, nil
	}
	return nil, fmt.Errorf("unknown property type")
}

func ReadNode(reader *bufio.Reader) (*Node, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	if b != 0 {
		return nil, fmt.Errorf("element must start with 0-byte")
	}
	nameIndex, err := ReadNumber(reader)
	if err != nil {
		return nil, err
	}
	numProps, err := ReadNumber(reader)
	if err != nil {
		return nil, err
	}
	numElems, err := ReadNumber(reader)
	if err != nil {
		return nil, err
	}
	node := &Node{Name: nameIndex.Value}
	for i := 0; i < numProps.Value; i++ {
		p, err := ReadProperty(reader)
		if err != nil {
			return nil, err
		}
		node.Properties = append(node.Properties, p)
	}
	for i := 0; i < numElems.Value; i++ {
		c, err := ReadNode(reader)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, c)
	}
	return node, nil
}

func (d *Document) GetTypeAndValue(val TypedValue, options *Options) (string, string) {
	switch val.(type) {
	case String:
		typeId := val.TypeId()
		if typeId >= 0xc0 && typeId < 0xd0 {
			v := typeId - 0xc0
			return "string", d.Strings[v]
		} else if typeId == 0xd0 || typeId == 0xe0 {
			v := val.(String)
			return "string", d.Strings[v.Value]
		}
	case Float:
		v := val.(Float)
		return "float", fmt.Sprintf("%f", v.Value)
	case Bool:
		v := val.(Bool)
		if v.Value {
			return "bool", "true"
		}
		return "bool", "false"
	case Int8:
		v := val.(Int8)
		if options.Hex8 {
			return "int8", fmt.Sprintf("0x%X", uint8(v.Value))
		}
		return "int8", fmt.Sprintf("%d", v.Value)
	case Int16:
		v := val.(Int16)
		if options.Hex16 {
			return "int16", fmt.Sprintf("0x%X", uint16(v.Value))
		}
		return "int16", fmt.Sprintf("%d", v.Value)
	case Int32:
		v := val.(Int32)
		if options.Hex32 {
			return "int32", fmt.Sprintf("0x%X", uint32(v.Value))
		}
		return "int32", fmt.Sprintf("%d", v.Value)
	}
	return "_?_", "_?_"
}

func (d *Document) WriteProperty(enc *xml.Encoder, prop *Property, options *Options) error {
	name := d.Strings[prop.Name]
	typ, val := d.GetTypeAndValue(d.TypedValues[prop.Value], options)

	t := xml.StartElement{
		Name: xml.Name{Local: "property"},
		Attr: []xml.Attr{
			xml.Attr{Name: xml.Name{Local: "name"}, Value: name},
			xml.Attr{Name: xml.Name{Local: "type"}, Value: typ},
			xml.Attr{Name: xml.Name{Local: "value"}, Value: val},
		},
	}
	err := enc.EncodeToken(t)
	if err != nil {
		return err
	}
	err = enc.EncodeToken(t.End())
	if err != nil {
		return err
	}
	return nil
}

func (d *Document) WriteNode(enc *xml.Encoder, node *Node, options *Options) error {
	// start tag
	name := d.Strings[node.Name]
	t := xml.StartElement{Name: xml.Name{Local: name}}
	err := enc.EncodeToken(t)
	if err != nil {
		return err
	}
	// child nodes
	for _, p := range node.Properties {
		d.WriteProperty(enc, p, options)
	}
	for _, c := range node.Children {
		d.WriteNode(enc, c, options)
	}
	// end tag
	err = enc.EncodeToken(t.End())
	if err != nil {
		return err
	}
	return nil
}
