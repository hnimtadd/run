build:
	@ go build -o ./bin ./cmd/main.go 

run: build
	@ ./bin/main

test: 
	@ go test -v ./...

gen:
	@ cd proto && buf generate

clean_example:
	@rm **.wasm

build_example:
	@GOOS=wasip1 GOARCH=wasm go build -o test/go/data/main.wasm test/go/data/main.go
	@GOOS=wasip1 GOARCH=wasm go build -o internal/_testdata/helloworld.wasm internal/_testdata/helloworld.go
	@GOOS=wasip1 GOARCH=wasm go build -o examples/go/example.wasm examples/go/example.go

go_lint:
	@golangci-lint run  ./...

.PHONY: build run gen test build_example clean_example go_lint
