package data

import (
	"fmt"
	"strings"
)

type Document struct {
	Signature      uint32
	NumStrings     Number
	Strings        []string
	NumTypedValues Number
	TypedValues    []TypedValue
	Element        *Node
}

type Number struct {
	Value uint
}

type TypedValue interface {
	TypeId() int
	Encode() []byte
}

type Node struct {
	Name       *string
	Properties []*Property
	Children   []*Node
}

type Property struct {
	Name  *string
	Value TypedValue
}

func (p Property) String() string {
	return fmt.Sprintf("Prop{%s %v}", *p.Name, p.Value)
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
	return fmt.Sprintf("Elem{%s, props:[%s], elems:[%s]}", *n.Name, props, elems)
}
