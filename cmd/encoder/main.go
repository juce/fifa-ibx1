package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"juce/fifa-ibx1/data"
	"os"
)

var Version = "unknown"

type PropList struct {
	props []XmlProp
}

type XmlProp struct {
	name  string
	typ   string
	value string
}

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <infile> <outfile> [options]\n", os.Args[0])
		os.Exit(0)
	}

	infile := os.Args[1]
	outfile := os.Args[2]
	fmt.Printf("converting %s --> %s ...\n", infile, outfile)

	f, err := os.Open(infile)
	if err != nil {
		fmt.Errorf("opening input file: %v", err)
		os.Exit(1)
	}
	defer f.Close()

	doc := data.Document{ShareTypedValues: true}

	dec := xml.NewDecoder(bufio.NewReader(f))

	var stack []*data.Node
	var propStack []*PropList

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
				x := XmlProp{}
				for _, a := range tok.Attr {
					//fmt.Printf("%#v = %#v\n", a.Name.Local, a.Value)
					if a.Name.Local == "name" {
						x.name = string(a.Value)
					} else if a.Name.Local == "type" {
						x.typ = a.Value
					} else if a.Name.Local == "value" {
						x.value = a.Value
					}
				}
				li := propStack[len(propStack)-1]
				li.props = append(li.props, x)
			} else {
				// element
				elem := &data.Node{Name: doc.GetString(tok.Name.Local)}
				elem.Properties = []*data.Property{}
				elem.Children = []*data.Node{}
				stack = append(stack, elem)                //push
				propStack = append(propStack, &PropList{}) //push
			}
		case xml.EndElement:
			if tok.Name.Local != "property" {
				elem := stack[len(stack)-1]
				//fmt.Printf("elem ending: %v\n", *elem.Name)
				stack = stack[:len(stack)-1] //pop
				// add props
				li := propStack[len(propStack)-1]
				propStack = propStack[:len(propStack)-1] //pop
				for _, x := range li.props {
					p := &data.Property{
						Name:  doc.GetString(x.name),
						Value: doc.GetTypedValue(x.typ, x.value),
					}
					elem.Properties = append(elem.Properties, p)
				}
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

	//fmt.Printf("structure: %v\n", doc)

	f, err = os.Create(outfile)
	if err != nil {
		fmt.Printf("opening output file: %v", err)
		os.Exit(1)
	}
	defer f.Close()

	result := doc.Encode()
	f.Write(result)
}
