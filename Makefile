all: encoder decoder

encoder: cmd/encoder/main.go data/tree.go data/types.go data/encoding.go
	go build -o encoder cmd/encoder/main.go

decoder: cmd/decoder/main.go data/tree.go data/types.go data/encoding.go
	go build -o decoder cmd/decoder/main.go
