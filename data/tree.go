package data

import (
	"fmt"
	"strconv"
	"strings"
)

type Document struct {
	Strings     []string
	TypedValues []TypedValue
	sMap        map[string]int
	tvMap       map[string]int
	Element     *Node
}

type Number struct {
	Value uint
}

type TypedValue interface {
	TypeId() int
	Encode() []byte
}

type Node struct {
	Name       int
	Properties []*Property
	Children   []*Node
}

type Property struct {
	Name  int
	Value int
}

func (p Property) String() string {
	return fmt.Sprintf("p{%v %v}", p.Name, p.Value)
}

func (n Node) String() string {
	arr := make([]string, len(n.Properties))
	for i, p := range n.Properties {
		arr[i] = fmt.Sprintf("%v", p)
	}
	props := strings.Join(arr, ",")
	arr = make([]string, len(n.Children))
	for i, c := range n.Children {
		arr[i] = fmt.Sprintf("%v", c)
	}
	elems := strings.Join(arr, ",")
	return fmt.Sprintf("E{%v, [%s], [%s]}", n.Name, props, elems)
}

func (d *Document) GetString(val string) int {
	index, ok := d.sMap[val]
	if ok {
		return index
	}
	index = len(d.Strings)
	d.Strings = append(d.Strings, val)
	if d.sMap == nil {
		d.sMap = make(map[string]int, 0)
	}
	d.sMap[val] = index
	return index
}

func (d *Document) GetTypedValue(typ string, val string) int {
	key := fmt.Sprintf("%s:%s", typ, val)
	index, ok := d.tvMap[key]
	if ok {
		//return index
	}
	index = len(d.TypedValues)
	var tv TypedValue
	if typ == "int8" {
		v, err := strconv.Atoi(val)
		if err == nil {
			tv = Int8{int8(v)}
		}
	} else if typ == "uint8" {
		v, err := strconv.Atoi(val)
		if err == nil {
			tv = UInt8{uint8(v)}
		}
	} else if typ == "int16" {
		v, err := strconv.Atoi(val)
		if err == nil {
			tv = Int16{int16(v)}
		}
	} else if typ == "uint16" {
		v, err := strconv.Atoi(val)
		if err == nil {
			tv = UInt16{uint16(v)}
		}
	} else if typ == "int32" {
		v, err := strconv.Atoi(val)
		if err == nil {
			tv = Int32{int32(v)}
		}
	} else if typ == "uint32" {
		v, err := strconv.Atoi(val)
		if err == nil {
			tv = UInt32{uint32(v)}
		}
	} else if typ == "string" {
		index := d.GetString(val)
		tv = String{index}
	} else if typ == "bool" {
		if val == "true" {
			tv = Bool{true}
		} else {
			tv = Bool{false}
		}
	} else if typ == "float" {
		v, err := strconv.ParseFloat(val, 32)
		if err == nil {
			tv = Float{float32(v)}
		}
	}
	d.TypedValues = append(d.TypedValues, tv)
	if d.tvMap == nil {
		d.tvMap = make(map[string]int, 0)
	}
	d.tvMap[key] = index
	return index
}
