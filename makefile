build:
	@ go build -o ./bin ./cmd/main.go 

run: build
	@ ./bin/main

.PHONY: build run
