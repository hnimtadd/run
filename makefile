PKG := github.com/hnimtadd/run
VERSION := $(shell git describe --always --long --dirty)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor | grep -v /docker)

BIN := ./bin
BUILD := `git describe --tags --abbrev=8 --dirty --always --long` 
LDFLAGS=-ldflags "-X=${PKG}/internal/version.Version=$(BUILD)"

clean:
	-@rm ${BIN}/*

build-ingress:
	@ go build ${LDFLAGS} -o ${BIN}/ingress ./cmd/ingress/main.go

ingress: build-ingress
	@ ${BIN}/ingress

buildapi:
	@ go build ${LDFLAGS} -o ${BIN}/api ./cmd/api/main.go

api: buildapi
	@ ${BIN}/api

test: 
	@ go clean -testcache
	@ go test --short ${PKG_LIST}

vet:
	@ go vet ${PKG_LIST}

gen:
	@ cd proto && buf generate

clean_example:
	@rm **/*.wasm

build_example:
	@GOOS=wasip1 GOARCH=wasm go build -o internal/_testdata/helloworld.wasm internal/_testdata/helloworld.go
	@GOOS=wasip1 GOARCH=wasm go build -o examples/go/example.wasm examples/go/example.go
	sh ./scripts/build_py_example.sh

golint:
	@golangci-lint run  ./...

containerup:
	@ docker-compose -f ./docker/docker-compose.yml --env-file ${ENV}.env.docker up -d --remove-orphans

containerdown:
	@ docker-compose -f ./docker/docker-compose.yml --env-file ${ENV}.env.docker down

.PHONY: build-ingress ingress build-api api gen test build_example clean_example go-lint container-up container-down
