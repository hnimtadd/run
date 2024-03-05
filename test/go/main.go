package main

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed data/main.wasm
var addWasm []byte

func main() {
	now := time.Now()
	flag.Parse()
	ctx := context.Background()

	r := wazero.NewRuntime(ctx)
	defer func() {
		if err := r.Close(ctx); err != nil {
			fmt.Printf(`cannot close, err %v`, err)
		}
	}()

	wasi_snapshot_preview1.MustInstantiate(ctx, r)
	mod, err := r.Instantiate(ctx, addWasm)
	if err != nil {
		log.Panicf("failed to instantiate module:%v", err.Error())
	}
	x, y, err := readTwoArgs(flag.Arg(0), flag.Arg(1))
	if err != nil {
		log.Panicf("failed to read argument: %v", err)
	}
	add := mod.ExportedFunction("add")
	results, err := add.Call(ctx, x, y)
	if err != nil {
		log.Panicf("failed to call add: %v", err.Error())
	}
	fmt.Printf("result: %d + %d = %d\n", x, y, results[0])
	fmt.Printf("running example in %fs\n", time.Since(now).Seconds())
}

func readTwoArgs(xs, ys string) (uint64, uint64, error) {
	if xs == "" || ys == "" {
		return 0, 0, errors.New("must specify two command line argument")
	}
	x, err := strconv.ParseUint(xs, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("argument X: %v", err)
	}

	y, err := strconv.ParseUint(ys, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("argument Y: %v", err)
	}

	return x, y, nil
}
