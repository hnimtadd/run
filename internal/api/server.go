package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hnimtadd/run/internal/settings"
	"github.com/hnimtadd/run/internal/store"
	"github.com/hnimtadd/run/internal/types"
	"github.com/hnimtadd/run/internal/utils"

	"github.com/go-chi/chi/v5"
)

// TODO: add get,update endpoint, upload deployment.

const _24k = (1 >> 10) * 24

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
	params := new(CreateEndpointParams)
	if err := json.NewDecoder(r.Body).Decode(params); err != nil {
		return utils.WriteJSON(w, http.StatusBadRequest, utils.MakeErrorResponse(err))
	}
	defer func() { _ = r.Body.Close() }()

	endpoint, err := types.NewEnpoint(params.Name, params.Runtime, params.Environment)
	if err != nil {
		return utils.WriteJSON(w, http.StatusBadRequest, utils.MakeErrorResponse(err))
	}
	if err := s.store.CreateEndpoint(endpoint); err != nil {
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}
	return utils.WriteJSON(w, http.StatusOK, endpoint)
}

func (s *Server) HandleGetEndpointByID(w http.ResponseWriter, r *http.Request) error {
	endpointID := chi.URLParam(r, "id")
	fmt.Println("get endpoint id", endpointID)

	endpoint, err := s.store.GetEndpointByID(endpointID)
	if err != nil {
		return utils.WriteJSON(w, http.StatusNotFound, utils.MakeErrorResponse(err))
	}

	return utils.WriteJSON(w, http.StatusOK, endpoint)
}

func (s *Server) HandleGetEndpoints(w http.ResponseWriter, _ *http.Request) error {
	endpoints, err := s.store.GetEndpoints()
	if err != nil {
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}

	return utils.WriteJSON(w, http.StatusOK, endpoints)
}

func (s *Server) HandlePostDeployment(w http.ResponseWriter, r *http.Request) error {
	endpointID := chi.URLParam(r, "id")
	endpoint, err := s.store.GetEndpointByID(endpointID)
	if err != nil {
		return utils.WriteJSON(w, http.StatusNotFound, utils.MakeErrorResponse(err))
	}

	if err := r.ParseMultipartForm(_24k); err != nil {
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}

	f, _, err := r.FormFile("blob")
	if err != nil {
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}

	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, f)
	if err != nil {
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}

	if size >= settings.MaxBlobSize {
		fmt.Println(size, settings.MaxBlobSize)
		return utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "given blob exceed maxsize"})
	}

	deployment, err := types.NewDeployment(endpoint, buf.Bytes())
	if err != nil {
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}

	if err := s.store.CreateDeployment(deployment); err != nil {
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}
	return utils.WriteJSON(w, http.StatusOK, deployment)
}

func (s *Server) HandleGetDeploymentsOfEndpoint(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

// HandleGetDeployment TODO: get deployment from store, exclude Blob
func (s *Server) HandleGetDeployment(_ http.ResponseWriter, _ *http.Request) error {
	return nil
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
	s.router.Get("/endpoint/{id}", makeAPIHandler(s.HandleGetEndpointByID))
	s.router.Post("/endpoint/{id}/deploy", makeAPIHandler(s.HandlePostDeployment))
	s.router.Get("/endpoint/{id}/deploy", makeAPIHandler(s.HandleGetDeploymentsOfEndpoint))
	s.router.Get("/deploy/{id}", makeAPIHandler(s.HandleGetDeployment))
}

func (s *Server) ListenAndServe(addr string) error {
	s.InitRoute()
	fmt.Printf("Listen and serve api at: %v\n", addr)
	return http.ListenAndServe(addr, s.router)
}
