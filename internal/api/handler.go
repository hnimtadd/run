package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type (
	apiHandler    func(w http.ResponseWriter, r *http.Request) error
	errorResponse struct {
		Error string `json:"error"`
	}
)

func makeErrorResponse(err error) errorResponse {
	return errorResponse{
		Error: err.Error(),
	}
}

func makeAPIHandler(h apiHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			// todo
			fmt.Printf("api handler error, %v", err)
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
