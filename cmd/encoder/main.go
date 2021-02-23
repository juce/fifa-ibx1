package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"juce/fifa-ibx1/data"
	"os"
	"path"
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
		fmt.Printf("FIFA IBX1 Encoder by juce. Version: %s\n", Version)
		fmt.Printf("Usage: %s <in-path> <out-path> [options]\n", os.Args[0])
		fmt.Printf("Options:\n")
		fmt.Printf("\t--debug   : print out extra info for troubleshooting\n")
		fmt.Printf("\t--noshare : do not re-use typed values (produces larger IBX1 files)\n")
		os.Exit(0)
	}

	infile := os.Args[1]
	outfile := os.Args[2]
	options := os.Args[3:]

	fi, err := os.Stat(infile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var count int
	if fi.IsDir() {
		// input is a directory
		count = ProcessDir(infile, outfile, options)
	} else {
		// check if output is an existing directory
		fi, err := os.Stat(outfile)
		if err == nil && fi.IsDir() {
			ext := path.Ext(infile)
			outfile = path.Join(outfile, fmt.Sprintf("%s%s", infile[:len(infile)-len(ext)], ".dat"))
		}
		count = ProcessFile(infile, outfile, options)
	}
	if count < 0 {
		os.Exit(1)
	}
	fmt.Println("files processed:", count)
}

func ProcessDir(indir string, outdir string, opts []string) int {
	count := 0
	entries, err := ioutil.ReadDir(indir)
	if err != nil {
		fmt.Printf("problem reading directory: %v\n", err)
		return -1
	}
	err = os.MkdirAll(outdir, 0775)
	if err != nil {
		fmt.Printf("problem creating output directory: %v\n", err)
		return -1
	}
	for _, entry := range entries {
		if entry.Name() == "." || entry.Name() == ".." {
			continue
		}
		inItem := path.Join(indir, entry.Name())
		outItem := path.Join(outdir, entry.Name())
		if entry.IsDir() {
			count += ProcessDir(inItem, outItem, opts)
		} else {
			ext := path.Ext(outItem)
			outItem = fmt.Sprintf("%s%s", outItem[:len(outItem)-len(ext)], ".dat")
			count += ProcessFile(inItem, outItem, opts)
		}
	}
	return count
}

func ProcessFile(infile string, outfile string, opts []string) int {
	fmt.Printf("converting %s --> %s ... ", infile, outfile)

	var options data.Options
	for _, opt := range opts {
		if opt == "--debug" {
			options.Debug = true
		} else if opt == "--noshare" {
			options.NoShare = true
		}
	}

	f, err := os.Open(infile)
	if err != nil {
		fmt.Errorf("opening input file: %v", err)
		return -1
	}
	defer f.Close()

	doc := data.Document{ShareTypedValues: !options.NoShare}

	dec := xml.NewDecoder(bufio.NewReader(f))

	var stack []*data.Node
	var propStack []*PropList

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return -1
		}
		stop := false
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
				if len(tok.Attr) > 0 {
					fmt.Printf("warn: element %s has attributes ", tok.Name.Local)
					stop = true
					break
				}
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
		if stop {
			break
		}
	}

	if doc.Element == nil {
		f.Seek(0, 0)
		reader := bufio.NewReader(f)

		// copy all data unmodified
		outf, err := os.Create(outfile)
		if err != nil {
			fmt.Printf("%v\n", err)
			return -1
		}
		defer outf.Close()

		chunk := make([]byte, 16384)
		for {
			n, err := io.ReadFull(reader, chunk)
			outf.Write(chunk[:n])
			if err == io.EOF {
				break
			}
		}
		fmt.Println("OK (unchanged)")
		return 1
	}

	if options.Debug {
		fmt.Printf("\n")
		fmt.Printf("num strings: 0x%x (%d)\n", len(doc.Strings), len(doc.Strings))
		for i, s := range doc.Strings {
			fmt.Printf("0x%x (%d): 0x%x {%s}\n", i, i, len(s), s)
		}
		fmt.Printf("num typed-values: 0x%x (%d)\n", len(doc.TypedValues), len(doc.TypedValues))
		for i, v := range doc.TypedValues {
			fmt.Printf("0x%x (%d): %v\n", i, i, v)
		}
		fmt.Printf("%v\n", doc)
	}

	f, err = os.Create(outfile)
	if err != nil {
		fmt.Printf("opening output file: %v", err)
		return -1
	}
	defer f.Close()

	result := doc.Encode()
	f.Write(result)
	fmt.Println("OK")
	return 1
}
