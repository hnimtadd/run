package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hnimtadd/run/internal/settings"
	"github.com/hnimtadd/run/internal/store"
	"github.com/hnimtadd/run/internal/types"
	"github.com/hnimtadd/run/internal/utils"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	store    store.Store
	logStore store.LogStore
	router   *chi.Mux
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

	endpoint, err := types.NewEndpoint(params.Name, params.Runtime, params.Environment)
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

	deployments, err := s.store.GetDeploymentByEndpointID(endpointID)
	if err != nil {
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}

	return utils.WriteJSON(w, http.StatusOK, FromInternalEndpoint(endpoint, deployments))
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

	if err := r.ParseMultipartForm(settings.MaxBlobSize); err != nil {
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

	// TODO: fix, currently, if user need to update new environment value to the request, we must extract it from the body.
	deployment, err := types.NewDeployment(endpoint, buf.Bytes(), endpoint.Environment)
	if err != nil {
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}

	if err := s.store.CreateDeployment(deployment); err != nil {
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}
	if err := s.store.UpdateEndpoint(endpoint.ID.String(), store.UpdateEndpointParams{ActiveDeployID: deployment.ID, Environment: deployment.Environment}); err != nil {
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}

	return utils.WriteJSON(w, http.StatusOK, FromInternalDeployment(deployment))
}

func (s *Server) HandleGetDeploymentsOfEndpoint(w http.ResponseWriter, r *http.Request) error {
	endpointID := chi.URLParam(r, "id")

	if _, err := s.store.GetEndpointByID(endpointID); err != nil {
		return utils.WriteJSON(w, http.StatusNotFound, utils.MakeErrorResponse(err))
	}

	deployments, err := s.store.GetDeploymentByEndpointID(endpointID)
	if err != nil {
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}
	var d []Deployment
	for _, deployment := range deployments {
		d = append(d, FromInternalDeployment(deployment))
	}
	return utils.WriteJSON(w, http.StatusOK, d)
}

func (s *Server) HandleGetDeployment(w http.ResponseWriter, r *http.Request) error {
	deploymentID := chi.URLParam(r, "id")

	deployment, err := s.store.GetDeploymentByID(deploymentID)
	if err != nil {
		return utils.WriteJSON(w, http.StatusNotFound, utils.MakeErrorResponse(err))
	}

	return utils.WriteJSON(w, http.StatusOK, FromInternalDeployment(deployment))
}

func (s *Server) HandleGetLogOfRequest(w http.ResponseWriter, r *http.Request) error {
	requestID := chi.URLParam(r, "id")
	log, err := s.logStore.GetLogByRequestID(requestID)
	if err != nil {
		return utils.WriteJSON(w, http.StatusNotFound, utils.MakeErrorResponse(err))
	}
	return utils.WriteJSON(w, http.StatusOK, FromInternalRequestLog(log))
}

func FromInternalRequestLog(log *types.RequestLog) map[string]any {
	return map[string]any{
		"requestID":    log.RequestID.String(),
		"deploymentID": log.RequestID.String(),
		"logs":         log.Contents,
		"createdAt":    time.Unix(log.CreatedAt, 0).String(),
	}
}

func (s *Server) HandleGetLogOfDeployment(w http.ResponseWriter, r *http.Request) error {
	deploymentID := chi.URLParam(r, "id")
	_, err := s.store.GetDeploymentByID(deploymentID)
	if err != nil {
		return utils.WriteJSON(w, http.StatusNotFound, utils.MakeErrorResponse(err))
	}
	logs, err := s.logStore.GetLogOfDeployment(deploymentID)
	if err != nil {
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}
	var rspLogs []map[string]any
	for _, log := range logs {
		rspLogs = append(rspLogs, FromInternalRequestLog(log))
	}

	return utils.WriteJSON(w, http.StatusOK, rspLogs)
}

func handleStatus(w http.ResponseWriter, _ *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	status := map[string]string{"status": "ok"}
	return json.NewEncoder(w).Encode(status)
}

func NewServer(store store.Store, logStore store.LogStore) *Server {
	return &Server{
		store:    store,
		logStore: logStore,
	}
}

func (s *Server) InitRoute() {
	s.router = chi.NewRouter()
	s.router.Get("/status", makeAPIHandler(handleStatus))
	s.router.Post("/endpoint", makeAPIHandler(s.HandleCreateEndpoint))
	s.router.Get("/endpoint/{id}", makeAPIHandler(s.HandleGetEndpointByID))
	s.router.Post("/endpoint/{id}/deploy", makeAPIHandler(s.HandlePostDeployment))
	s.router.Get("/endpoint/{id}/deploy", makeAPIHandler(s.HandleGetDeploymentsOfEndpoint))
	s.router.Get("/deployment/{id}", makeAPIHandler(s.HandleGetDeployment))
	s.router.Get("/deployment/{id}/log", makeAPIHandler(s.HandleGetLogOfDeployment))

	s.router.Get("/request/{id}/log", makeAPIHandler(s.HandleGetLogOfRequest))
}

func (s *Server) ListenAndServe(addr string) error {
	s.InitRoute()
	fmt.Printf("Listen and serve api at: %v\n", addr)
	return http.ListenAndServe(addr, s.router)
}

type Deployment struct {
	ID         string `json:"id"`
	Hash       string `json:"hash"`
	EndpointID string `json:"endpointID"`
	Created    string `json:"createdAt"`
}

func FromInternalDeployment(d *types.Deployment) Deployment {
	return Deployment{
		ID:         d.ID.String(),
		Hash:       d.Hash,
		EndpointID: d.EndpointID.String(),
		Created:    time.Unix(d.CreatedAt, 0).String(),
	}
}

type Endpoint struct {
	ID                 string            `json:"id,omitempty"`
	Name               string            `json:"name,omitempty"`
	Runtime            string            `json:"runtime,omitempty"`
	Environment        map[string]string `json:"environment,omitempty"`
	ActiveDeploymentID string            `json:"activeDeploymentID,omitempty"`
	DeployHistory      []Deployment      `json:"deployHistory,omitempty"`
	CreatedAt          string            `json:"createdAt,omitempty"`
}

func FromInternalEndpoint(endpoint *types.Endpoint, deployments []*types.Deployment) Endpoint {
	var deployHistory []Deployment
	for _, deployment := range deployments {
		deployHistory = append(deployHistory, FromInternalDeployment(deployment))
	}
	return Endpoint{
		ID:                 endpoint.ID.String(),
		Name:               endpoint.Name,
		Runtime:            endpoint.Runtime,
		Environment:        endpoint.Environment,
		ActiveDeploymentID: endpoint.ActiveDeploymentID.String(),
		DeployHistory:      deployHistory,
		CreatedAt:          time.Unix(endpoint.CreatedAt, 0).String(),
	}
}
