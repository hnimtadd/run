package api

import (
	"encoding/json"
	"net/http"

	"github.com/hnimtadd/run/internal/store"
	"github.com/hnimtadd/run/internal/types"

	"github.com/go-chi/chi/v5"
)

// TODO: add get,update endpoint, upload deployment.

type Server struct {
	store  store.Store
	router *chi.Mux
}

type CreateEndpointParams struct {
	Name        string            `json:"name"`        // Name of the endpoint
	Runtime     string            `json:"runtime"`     // Runtime on which the code will be invoked. (go or js for now)
	Environment map[string]string `json:"environment"` // A map of environment variables
}

func (s *Server) HandleCreateEndpoint(w http.ResponseWriter, r *http.Request) error {
	var params *CreateEndpointParams
	if err := json.NewDecoder(r.Body).Decode(params); err != nil {
		return writeJSON(w, http.StatusBadRequest, makeErrorResponse(err))
	}
	defer func() { _ = r.Body.Close() }()

	endpoint, err := types.NewEnpoint(params.Name, params.Runtime, params.Environment)
	if err != nil {
		return writeJSON(w, http.StatusBadRequest, makeErrorResponse(err))
	}
	if err := s.store.CreateEndpoint(endpoint); err != nil {
		return writeJSON(w, http.StatusInternalServerError, makeErrorResponse(err))
	}
	return writeJSON(w, http.StatusOK, endpoint)
}

func handleStatus(w http.ResponseWriter, _ *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	status := map[string]string{"status": "ok"}
	return json.NewEncoder(w).Encode(status)
}

func NewServer(store store.Store) *Server {
	return &Server{
		store: store,
	}
}

func (s *Server) InitRoute() {
	s.router = chi.NewRouter()
	s.router.Get("/status", makeAPIHandler(handleStatus))
	s.router.Post("/endpoint", makeAPIHandler(s.HandleCreateEndpoint))
}

func (s *Server) ListenAndServe(addr string) error {
	s.InitRoute()
	return http.ListenAndServe(addr, s.router)
}
