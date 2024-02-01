build:
	@ go build -o ./bin ./cmd/main.go 

run: build
	@ ./bin/main

gen:
	@ cd proto && buf generate

.PHONY: build run gen
