package main

import (
	"net/http"

	"github.com/hnimtadd/run/sdk"
)

func handle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello world!"))
}

func main() {
	sdk.Handle(http.HandlerFunc(handle))
}
