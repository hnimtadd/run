PKG := github.com/hnimtadd/run
VERSION := $(shell git describe --always --long --dirty)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)
BIN := ./bin
BUILD := `git describe --tags --abbrev=8 --dirty --always --long` 
LDFLAGS=-ldflags "-X=${PKG}/internal/version.Version=$(BUILD)"

clean:
	-@rm ${BIN}/*

build-ingress:
	@ go build ${LDFLAGS} -o ${BIN}/ingress ./cmd/ingress/main.go

ingress: build-ingress
	@ ${BIN}/ingress

build-api:
	@ go build ${LDFLAGS} -o ${BIN}/api ./cmd/api/main.go

api: build-api
	@ ${BIN}/api

test: 
	@ go test --short ${PKG_LIST}

vet:
	@ go vet ${PKG_LIST}

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

.PHONY: build-ingress ingress build-api api gen test build_example clean_example go-lint container-up container-down
