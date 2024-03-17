package api

import (
	"fmt"
	"net/http"
)

type (
	apiHandler func(w http.ResponseWriter, r *http.Request) error
)

func makeAPIHandler(h apiHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			fmt.Printf("api handler error, %v", err)
		}
	}
}
