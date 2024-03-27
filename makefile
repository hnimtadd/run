build:
	@ go build -o ./bin/local ./cmd/main.go

run: build
	@ ./bin/local

build-ingress:
	@ go build -o ./bin/ingress ./cmd/ingress/main.go

ingress: build-ingress
	@ ./bin/ingress

build-api:
	@ go build -o ./bin/api ./cmd/api/main.go

api: build-api
	@ ./bin/api

test: 
	@ go test -v ./...

gen:
	@ cd proto && buf generate

clean_example:
	@rm **.wasm

build_example:
	@GOOS=wasip1 GOARCH=wasm go build -o internal/_testdata/helloworld.wasm internal/_testdata/helloworld.go
	@GOOS=wasip1 GOARCH=wasm go build -o examples/go/example.wasm examples/go/example.go

go-lint:
	@golangci-lint run  ./...

container-up:
	@ docker-compose -f ./docker/docker-compose.yml up -d

container-down:
	@ docker-compose -f ./docker/docker-compose.yml down

.PHONY: build run build-ingress ingress build-api api gen test build_example clean_example go-lint container-up container-down
