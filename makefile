PKG := github.com/hnimtadd/run
VERSION := $(shell git describe --always --long --dirty)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor | grep -v /docker)

BIN := ./bin
BUILD := `git describe --tags --abbrev=8 --dirty --always --long` 
LDFLAGS=-ldflags "-X=${PKG}/internal/version.Version=$(BUILD)"

clean:
	-@rm ${BIN}/*

buildingress:
	@ go build ${LDFLAGS} -o ${BIN}/ingress ./cmd/ingress/main.go

ingress: buildingress
	@ ${BIN}/ingress

buildapi:
	@ go build ${LDFLAGS} -o ${BIN}/api ./cmd/api/main.go

api: buildapi
	@ ${BIN}/api

test:  build_example
	@ go clean -testcache
	@ go test --short ${PKG_LIST}

vet:
	@ go vet ${PKG_LIST}

gen:
	@ cd proto && buf generate

clean_example:
	@rm **/*.wasm

build_example:
	@GOOS=wasip1 GOARCH=wasm go build -o internal/_testdata/go/helloworld.wasm internal/_testdata/go/helloworld.go
	@GOOS=wasip1 GOARCH=wasm go build -o examples/go/example.wasm examples/go/example.go

build_py_example:
	sh ./scripts/build_py_example.sh

golint:
	@golangci-lint run  ./...

containerup:
	@ docker-compose -f ./docker/docker-compose.yml --env-file ${ENV}.env.docker up -d --remove-orphans

containerdown:
	@ docker-compose -f ./docker/docker-compose.yml --env-file ${ENV}.env.docker down

.PHONY: buildingress ingress buildapi api gen test build_example clean_example golint containerup containerdown build_py_example
