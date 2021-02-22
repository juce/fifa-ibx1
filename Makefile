all: xml2dat dat2xml

xml2dat: cmd/encoder/main.go data/tree.go data/types.go data/encoding.go
	go build -o xml2dat cmd/encoder/main.go

dat2xml: cmd/decoder/main.go data/tree.go data/types.go data/encoding.go
	go build -o dat2xml cmd/decoder/main.go
