package main

import (
	"fmt"
	"net/http"

	sdk "github.com/hnimtadd/run/sdk/go"
)

func handle(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello world!"))
	fmt.Println("hello, this is a request_log.go")
}

//export _start
func main() {
	sdk.Handle(http.HandlerFunc(handle))
}
