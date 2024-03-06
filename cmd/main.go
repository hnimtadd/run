package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hnimtadd/run/internal/actrs"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}

	system := actor.NewActorSystem()
	defer system.Shutdown()
	s := actrs.NewServer(os.Getenv("WASM_ADDR"))
	wasmPID := system.Root.Spawn(actor.PropsFromProducer(s))
	fmt.Println(wasmPID)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	<-c
}
