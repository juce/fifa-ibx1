VERSION=1.0
GIT_COMMIT=$(shell git -C . rev-parse HEAD)

all: xml2dat dat2xml

xml2dat: cmd/encoder/main.go data/tree.go data/types.go data/encoding.go
	go build -ldflags="-X main.Version=${VERSION}-${GIT_COMMIT}" -o xml2dat cmd/encoder/main.go

dat2xml: cmd/decoder/main.go data/tree.go data/types.go data/encoding.go
	go build -ldflags="-X main.Version=${VERSION}-${GIT_COMMIT}" -o dat2xml cmd/decoder/main.go
