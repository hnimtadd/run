package main

import (
	"net/http"

	"github.com/hnimtadd/run/sdk"
)

func handle(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("Hello world!"))
	if err != nil {
		panic(err)
	}
}

func main() {
	sdk.Handle(http.HandlerFunc(handle))
}
