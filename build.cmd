@SET VERSION=1.2
@git -C . rev-parse HEAD >temp
@SET /p GIT_COMMIT= <temp

go build -ldflags="-X main.Version=%VERSION%-%GIT_COMMIT%" -o xml2dat.exe cmd/encoder/main.go
go build -ldflags="-X main.Version=%VERSION%-%GIT_COMMIT%" -o xml2dat.exe cmd/encoder/main.go

@del /Q temp
