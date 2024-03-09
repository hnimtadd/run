build:
	@ go build -o ./bin ./cmd/main.go 

run: build
	@ ./bin/main

test: 
	@ go run test/go/main.go  5 10
gen:
	@ cd proto && buf generate

build_example:
	@GOOS=wasip1 GOARCH=wasm go build -o examples/go/example.wasm examples/go/example.go
	@GOOS=wasip1 GOARCH=wasm go build -o internal/_testdata/helloworld.wasm internal/_testdata/helloworld.go

go_lint:
	@golangci-lint run  ./...

.PHONY: build run gen test build_example go_lint
