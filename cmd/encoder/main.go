package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"juce/fifa-ibx1/data"
	"os"
	"strconv"
)

var Version = "unknown"

func main() {
	/**
	strings := []string{
		"",
		"FIFA_Base",
	}
	val := data.Int16{4}
	val1 := data.UInt8{0x36}
	val2 := data.Int32{-1}
	val3 := data.String{1}
	val4 := data.String{134}
	val5 := data.Float{}
	val6 := data.Bool{true}
	val7 := data.Bool{}

	val8 := data.TypedValue(data.Int16{5})
	val9 := data.Int8{7}
	valA := data.String{12345}

	fmt.Printf("FIFA IBX1 decoder. Version %s\n", Version)
	fmt.Printf("val: %v\n", val)
	fmt.Printf("val1: %v\n", val1)
	fmt.Printf("val2: %v\n", val2)
	fmt.Printf("val3: %v\n", val3)
	fmt.Printf("val3: %#v\n", val3.Encode())
	fmt.Printf("val3: %s\n", val3.Deref(strings))
	fmt.Printf("val4: %v\n", val4)
	fmt.Printf("val4: %#v\n", val4.Encode())
	fmt.Printf("val5: %v\n", val5)
	fmt.Printf("val6: %v\n", val6)
	fmt.Printf("val7: %v\n", val7)
	fmt.Printf("val8: %v\n", val8)
	fmt.Printf("val8: %#v\n", val8.Encode())
	fmt.Printf("val9: %v\n", val9)
	fmt.Printf("val9: %#v\n", val9.Encode())
	fmt.Printf("valA: %#v\n", valA.Encode())


	fmt.Printf("bytes: %#v\n", data.Number{0x33}.Encode())
	fmt.Printf("bytes: %#v\n", data.Number{0xe7}.Encode())
	fmt.Printf("bytes: %#v\n", data.Number{0x1e3}.Encode())
	**/

	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <infile> <outfile> [options]\n", os.Args[0])
		os.Exit(0)
	}

	infile := os.Args[1]
	outfile := os.Args[2]
	fmt.Printf("converting %s --> %s ...\n", infile, outfile)

	f, err := os.Open(infile)
	if err != nil {
		fmt.Errorf("opening file: %v", err)
		return
	}
	defer f.Close()

	var doc data.Document
	dec := xml.NewDecoder(bufio.NewReader(f))
	//err = dec.Decode(&doc.Element)
	//if err != nil {
	//	fmt.Errorf("reading xml: %v", err)
	//}

	var stack []*data.Node
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			os.Exit(1)
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			//fmt.Printf("%s\n", strings.Join(stack, " "))
			if tok.Name.Local == "property" {
				prop := &data.Property{}
				var typ string
				var val string
				for _, a := range tok.Attr {
					//fmt.Printf("%#v = %#v\n", a.Name.Local, a.Value)
					if a.Name.Local == "name" {
						name := string(a.Value)
						prop.Name = &name
					} else if a.Name.Local == "type" {
						typ = a.Value
					} else if a.Name.Local == "value" {
						val = a.Value
					}
				}
				if typ == "int8" {
					v, err := strconv.Atoi(val)
					if err == nil {
						prop.Value = data.Int8{int8(v)}
					}
				} else if typ == "uint8" {
					v, err := strconv.Atoi(val)
					if err == nil {
						prop.Value = data.UInt8{uint8(v)}
					}
				} else if typ == "int16" {
					v, err := strconv.Atoi(val)
					if err == nil {
						prop.Value = data.Int16{int16(v)}
					}
				} else if typ == "uint16" {
					v, err := strconv.Atoi(val)
					if err == nil {
						prop.Value = data.UInt16{uint16(v)}
					}
				} else if typ == "int32" {
					v, err := strconv.Atoi(val)
					if err == nil {
						prop.Value = data.Int32{int32(v)}
					}
				} else if typ == "uint32" {
					v, err := strconv.Atoi(val)
					if err == nil {
						prop.Value = data.UInt32{uint32(v)}
					}
				} else if typ == "string" {
					prop.Value = data.String{0}
				} else if typ == "bool" {
					if val == "true" {
						prop.Value = data.Bool{true}
					} else {
						prop.Value = data.Bool{false}
					}
				} else if typ == "float" {
					v, err := strconv.ParseFloat(val, 32)
					if err == nil {
						prop.Value = data.Float{float32(v)}
					}
				}

				// add property to parent element
				parent := stack[len(stack)-1]
				parent.Properties = append(parent.Properties, prop)
			} else {
				elem := &data.Node{Name: &tok.Name.Local}
				elem.Properties = []*data.Property{}
				elem.Children = []*data.Node{}
				stack = append(stack, elem) //push
			}
		case xml.EndElement:
			if tok.Name.Local != "property" {
				elem := stack[len(stack)-1]
				//fmt.Printf("elem ending: %v\n", *elem.Name)
				stack = stack[:len(stack)-1] //pop
				// add element to parent element, if parent exists
				if len(stack) > 0 {
					parent := stack[len(stack)-1]
					parent.Children = append(parent.Children, elem)
				} else {
					// done: assign
					doc.Element = elem
				}
			}
		}
	}

	fmt.Printf("structure: %v\n", doc)
}
