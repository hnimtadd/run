package main

import (
	"net/http"

	"github.com/hnimtadd/run/sdk"
)

func handle(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello world!"))
}

func main() {
	sdk.Handle(http.HandlerFunc(handle))
}
