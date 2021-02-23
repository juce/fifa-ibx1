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

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("FIFA IBX1 Decoder by juce. Version: %s\n", Version)
		fmt.Printf("Usage: %s <infile|indir> <outfile|outdir> [options]\n", os.Args[0])
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
			outfile = path.Join(outfile, fmt.Sprintf("%s%s", infile[:len(infile)-len(ext)], ".xml"))
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
		fmt.Println("problem reading directory: %v", err)
		return -1
	}
	err = os.MkdirAll(outdir, 0775)
	if err != nil {
		fmt.Println("problem creating output directory: %v", err)
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
			outItem = fmt.Sprintf("%s%s", outItem[:len(outItem)-len(ext)], ".xml")
			count += ProcessFile(inItem, outItem, opts)
		}
	}
	return count
}

func ProcessFile(infile string, outfile string, opts []string) int {
	fmt.Printf("converting %s --> %s ... ", infile, outfile)

	var options data.Options
	for _, opt := range opts {
		if opt == "--hex8" {
			options.Hex8 = true
		} else if opt == "--hex16" {
			options.Hex16 = true
		} else if opt == "--hex32" {
			options.Hex32 = true
		}
	}

	f, err := os.Open(infile)
	if err != nil {
		fmt.Printf("%v\n", err)
		return -1
	}
	defer f.Close()

	doc := data.Document{}

	reader := bufio.NewReader(f)

	sig := make([]byte, 4)
	_, err = io.ReadFull(reader, sig)
	if err != nil {
		fmt.Printf("reading signature: %v\n", err)
		return -1
	}
	if string(sig) != "IBX1" {
		// copy all data unmodified
		outf, err := os.Create(outfile)
		if err != nil {
			fmt.Printf("%v\n", err)
			return -1
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
		fmt.Println("OK (unchanged)")
		return 1
	}

	// seems to be an IBX1 file
	// num strings
	numStrings, err := data.ReadNumber(reader)
	if err != nil {
		fmt.Printf("reading number of strings: %v\n", err)
		return -1
	}
	//fmt.Printf("number of strings: 0x%x\n", numStrings.Value)
	// strings
	for i := 0; i < numStrings.Value; i++ {
		n, err := data.ReadNumber(reader)
		if err != nil {
			fmt.Printf("reading string length: %v\n", err)
			return -1
		}
		bs := make([]byte, n.Value)
		_, err = io.ReadFull(reader, bs)
		if err != nil {
			fmt.Printf("reading string: %v\n", err)
			return -1
		}
		_, err = reader.ReadByte() // 0-terminator
		if err != nil {
			fmt.Printf("reading string 0-terminator: %v\n", err)
			return -1
		}

		//fmt.Printf("0x%x 0x%x {%s}\n", i, n.Value, string(bs))
		doc.Strings = append(doc.Strings, string(bs))
	}
	// num typed values
	numTypedValues, err := data.ReadNumber(reader)
	if err != nil {
		fmt.Errorf("reading number of typed values: %v\n", err)
		return -1
	}
	//fmt.Printf("number of typed values: 0x%x\n", numTypedValues.Value)
	// typed values
	for i := 0; i < numTypedValues.Value; i++ {
		tv, err := data.ReadTypedValue(reader)
		if err != nil {
			fmt.Printf("reading typed value: %v\n", err)
			return -1
		}

		//fmt.Printf("0x%x %v\n", i, tv)
		doc.TypedValues = append(doc.TypedValues, tv)
	}
	// encoding flag
	_, err = reader.ReadByte()
	if err != nil {
		fmt.Printf("reading encoding flag: %v\n", err)
		return -1
	}
	// node structure
	node, err := data.ReadNode(reader)
	if err != nil {
		fmt.Printf("reading node structure: %v\n", err)
		return -1
	}
	doc.Element = node

	//fmt.Printf("%v\n", doc)

	// output as XML
	outf, err := os.Create(outfile)
	if err != nil {
		fmt.Printf("%v\n", err)
		return -1
	}
	defer outf.Close()

	outf.Write([]byte("<?xml version=\"1.0\" ?>\n"))
	writer := bufio.NewWriter(outf)

	enc := xml.NewEncoder(writer)
	enc.Indent("", "  ")
	err = doc.WriteNode(enc, doc.Element, &options)
	if err != nil {
		fmt.Printf("%v\n", err)
		return -1
	}
	writer.Flush()
	fmt.Println("OK")
	return 1
}
