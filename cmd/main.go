package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hnimtadd/run/internal/actrs"
	"github.com/hnimtadd/run/internal/api"
	"github.com/hnimtadd/run/internal/store"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}

	system := actor.NewActorSystem()
	defer system.Shutdown()
	st := store.NewMemoryStore()
	s := actrs.NewServer(fmt.Sprintf(":%v", os.Getenv("WASM_ADDR")), st)
	wasmPID := system.Root.Spawn(actor.PropsFromProducer(s))
	fmt.Println(wasmPID)
	apiServer := api.NewServer(st)
	go func() {
		panic(apiServer.ListenAndServe(fmt.Sprintf(":%v", os.Getenv("API_ADDR"))))
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	<-c
}
