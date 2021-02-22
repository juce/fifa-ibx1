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
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	doc := data.Document{}

	reader := bufio.NewReader(f)

	sig := make([]byte, 4)
	_, err = io.ReadFull(reader, sig)
	if err != nil {
		fmt.Printf("reading signature: %v\n", err)
		os.Exit(1)
	}
	if string(sig) != "IBX1" {
		// copy all data unmodified
		outf, err := os.Create(outfile)
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		defer outf.Close()

		outf.Write(sig)
		chunk := make([]byte, 16384)
		for {
			n, err := io.ReadFull(reader, chunk)
			outf.Write(chunk[:n])
			if err == io.EOF {
				break
			}
		}
		return
	}

	// seems to be an IBX1 file
	// num strings
	numStrings, err := data.ReadNumber(reader)
	if err != nil {
		fmt.Printf("reading number of strings: %v\n", err)
		os.Exit(1)
	}
	//fmt.Printf("number of strings: 0x%x\n", numStrings.Value)
	// strings
	for i := 0; i < numStrings.Value; i++ {
		n, err := data.ReadNumber(reader)
		if err != nil {
			fmt.Printf("reading string length: %v\n", err)
			os.Exit(1)
		}
		bs := make([]byte, n.Value)
		_, err = io.ReadFull(reader, bs)
		if err != nil {
			fmt.Printf("reading string: %v\n", err)
			os.Exit(1)
		}
		_, err = reader.ReadByte() // 0-terminator
		if err != nil {
			fmt.Printf("reading string 0-terminator: %v\n", err)
			os.Exit(1)
		}

		//fmt.Printf("0x%x 0x%x {%s}\n", i, n.Value, string(bs))
		doc.Strings = append(doc.Strings, string(bs))
	}
	// num typed values
	numTypedValues, err := data.ReadNumber(reader)
	if err != nil {
		fmt.Errorf("reading number of typed values: %v\n", err)
		os.Exit(1)
	}
	//fmt.Printf("number of typed values: 0x%x\n", numTypedValues.Value)
	// typed values
	for i := 0; i < numTypedValues.Value; i++ {
		tv, err := data.ReadTypedValue(reader)
		if err != nil {
			fmt.Printf("reading typed value: %v\n", err)
			os.Exit(1)
		}

		//fmt.Printf("0x%x %v\n", i, tv)
		doc.TypedValues = append(doc.TypedValues, tv)
	}
	// encoding flag
	_, err = reader.ReadByte()
	if err != nil {
		fmt.Printf("reading encoding flag: %v\n", err)
		os.Exit(1)
	}
	// node structure
	node, err := data.ReadNode(reader)
	if err != nil {
		fmt.Printf("reading node structure: %v\n", err)
		os.Exit(1)
	}
	doc.Element = node

	//fmt.Printf("%v\n", doc)

	// output as XML
	outf, err := os.Create(outfile)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	defer outf.Close()

	outf.Write([]byte("<?xml version=\"1.0\" ?>\n"))
	writer := bufio.NewWriter(outf)

	enc := xml.NewEncoder(writer)
	enc.Indent("", "  ")
	err = doc.WriteNode(enc, doc.Element)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	writer.Flush()
}
