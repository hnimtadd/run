build:
	@ go build -o ./bin ./cmd/main.go 

run: build
	@ ./bin/main

test: 
	@ go run test/go/main.go  5 10
gen:
	@ cd proto && buf generate

.PHONY: build run gen test
