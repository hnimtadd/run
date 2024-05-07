package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/hnimtadd/run/internal/settings"
	"github.com/hnimtadd/run/internal/store"
	"github.com/hnimtadd/run/internal/types"
	"github.com/hnimtadd/run/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// TODO: blob metadata related apis
type (
	Server struct {
		metadataStore store.Store
		blobStore     store.BlobStore
		logStore      store.LogStore
		router        *chi.Mux
		ServerConfig
	}
	ServerConfig struct {
		Addr    string
		Version string
	}
)

func NewServer(store store.Store, logStore store.LogStore, blobStore store.BlobStore, config ServerConfig) *Server {
	return &Server{
		metadataStore: store,
		logStore:      logStore,
		blobStore:     blobStore,
		ServerConfig:  config,
	}
}

func (s *Server) InitRoute() {
	s.router = chi.NewRouter()
	s.router.Get("/status", makeAPIHandler(handleStatus))
	s.router.Post("/endpoint", makeAPIHandler(s.HandleCreateEndpoint))
	s.router.Get("/endpoint/{id}", makeAPIHandler(s.HandleGetEndpointByID))
	s.router.Post("/endpoint/{id}/deploy", makeAPIHandler(s.HandlePostDeployment))
	s.router.Get("/endpoint/{id}/deploy", makeAPIHandler(s.HandleGetDeploymentsOfEndpoint))
	s.router.Options("/endpoint/{id}/rollback", makeAPIHandler(s.HandleRollback))

	s.router.Get("/deployment/{id}", makeAPIHandler(s.HandleGetDeployment))
	s.router.Get("/deployment/{id}/log", makeAPIHandler(s.HandleGetLogOfDeployment))

	s.router.Get("/request/{id}/log", makeAPIHandler(s.HandleGetLogOfRequest))
}

func (s *Server) ListenAndServe() error {
	s.InitRoute()
	slog.Info("Listen and serve api\n", "at", s.Addr, "version", s.Version)
	return http.ListenAndServe(s.Addr, s.router)
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
	if err := s.metadataStore.CreateEndpoint(endpoint); err != nil {
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}
	return utils.WriteJSON(w, http.StatusOK, endpoint)
}

func (s *Server) HandleGetEndpointByID(w http.ResponseWriter, r *http.Request) error {
	endpointID := chi.URLParam(r, "id")
	slog.Info("receive get endpoint by Id request", "endpointID", endpointID)
	endpoint, err := s.metadataStore.GetEndpointByID(endpointID)
	if err != nil {
		return utils.WriteJSON(w, http.StatusNotFound, utils.MakeErrorResponse(err))
	}

	deployments, err := s.metadataStore.GetDeploymentsByEndpointID(endpointID)
	if err != nil {
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}

	return utils.WriteJSON(w, http.StatusOK, FromInternalEndpoint(endpoint, deployments))
}

func (s *Server) HandleGetEndpoints(w http.ResponseWriter, _ *http.Request) error {
	endpoints, err := s.metadataStore.GetEndpoints()
	if err != nil {
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}

	return utils.WriteJSON(w, http.StatusOK, endpoints)
}

func (s *Server) HandlePostDeployment(w http.ResponseWriter, r *http.Request) error {
	endpointID := chi.URLParam(r, "id")
	endpoint, err := s.metadataStore.GetEndpointByID(endpointID)
	if err != nil {
		return utils.WriteJSON(w, http.StatusNotFound, utils.MakeErrorResponse(err))
	}

	if err := r.ParseMultipartForm(settings.MaxBlobSize); err != nil {
		slog.Info("cannot parse form of request ", "msg", err.Error())
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}

	f, _, err := r.FormFile("blob")
	if err != nil {
		slog.Info("cannot get file from form ", "msg", err.Error())
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}

	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, f)
	if err != nil {
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}

	if size >= settings.MaxBlobSize {
		slog.Info("request have blob exceed maxsize", "accept", settings.MaxBlobSize, "got", size)
		return utils.WriteJSON(w,
			http.StatusBadRequest,
			map[string]any{"error": "given blob exceed maxsize", "accepted": settings.MaxBlobSize})
	}

	// TODO: fix, currently, if user need to update new environment value to the request, we must extract it from the body.
	deployment, _ := types.NewDeployment(endpoint, buf.Bytes(), endpoint.Environment)

	blobMetadata, _ := types.NewRawBlobMetadata(deployment, buf.Bytes())

	_, err = s.blobStore.AddDeploymentBlob(blobMetadata, buf.Bytes())
	if err != nil {
		slog.Info("failed to store deployment blob", "msg", err.Error(), "node", "api server")
		return utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to create deployment",
		})
	}

	if err := s.metadataStore.CreateDeployment(deployment); err != nil {
		slog.Info("cannot create deployment in store", "msg", err.Error())
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}

	if err := s.metadataStore.UpdateActiveDeploymentOfEndpoint(endpoint.ID.String(), deployment.ID.String()); err != nil {
		slog.Info("cannot update active deployment for given endpoint", "msg", err.Error())
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}

	err = s.metadataStore.CreateBlobMetadata(blobMetadata)
	if err != nil {
		slog.Info("cannot create blob metadata", "msg", err.Error())
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}

	return utils.WriteJSON(w, http.StatusOK, FromInternalDeployment(deployment))
}

func (s *Server) HandleGetDeploymentsOfEndpoint(w http.ResponseWriter, r *http.Request) error {
	endpointID := chi.URLParam(r, "id")

	if _, err := s.metadataStore.GetEndpointByID(endpointID); err != nil {
		return utils.WriteJSON(w, http.StatusNotFound, utils.MakeErrorResponse(err))
	}

	deployments, err := s.metadataStore.GetDeploymentsByEndpointID(endpointID)
	if err != nil {
		return utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
	}
	var d []map[string]string
	for _, deployment := range deployments {
		d = append(d, FromInternalDeployment(deployment))
	}
	return utils.WriteJSON(w, http.StatusOK, d)
}

func (s *Server) HandleGetDeployment(w http.ResponseWriter, r *http.Request) error {
	deploymentID := chi.URLParam(r, "id")

	deployment, err := s.metadataStore.GetDeploymentByID(deploymentID)
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
	_, err := s.metadataStore.GetDeploymentByID(deploymentID)
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

func (s *Server) HandleRollback(w http.ResponseWriter, r *http.Request) error {
	endpointID := chi.URLParam(r, "id")
	deploymentID := r.URL.Query().Get("deploymentID")
	slog.Info("rollback for", "deploymentID", deploymentID, "endpoint", endpointID)

	endpoint, err := s.metadataStore.GetEndpointByID(endpointID)
	if err != nil {
		slog.Info("cannot get endpoint", "msg", err.Error())
		return utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "endpoint not existed"})
	}

	if !endpoint.HasActiveDeploy() {
		return utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot rollback on empty endpoint"})
	}

	deployments, err := s.metadataStore.GetDeploymentsByEndpointID(endpointID)
	if err != nil {
		slog.Info("cannot get deployments", "msg", err.Error())
		return utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot get deployments of given endpoint"})
	}

	switch deploymentID {
	case "":
		// default, rollback to latest deployment
		// maybe if endpoint have only 1 deployment, we could set the endpoint activeDeploymentID to null
		latestDeploymentUID := endpoint.ActiveDeploymentID
		currentDeploymentUID := uuid.Nil
		if len(deployments) >= 2 {
			currentDeploymentUID = deployments[len(deployments)-2].ID
		}

		blobMetadata, err := s.metadataStore.GetBlobMetadataByDeploymentID(latestDeploymentUID.String())
		if err != nil {
			slog.Info("cannot get blob metadata", "endpoint", endpointID, "deployment", latestDeploymentUID.String(), "msg", err.Error())
			return utils.WriteJSON(w, http.StatusBadRequest, utils.MakeErrorResponse(err))
		}
		deleted, err := s.blobStore.DeleteDeploymentBlob(blobMetadata.Location)
		if err != nil {
			slog.Info("cannot remove blob from blob storage", "endpoint", endpointID, "deployment", latestDeploymentUID.String(), "msg", err.Error())
			return utils.WriteJSON(w, http.StatusBadRequest, utils.MakeErrorResponse(err))
		}

		if !deleted {
			slog.Info("delete blob failed", "endpoint", endpointID, "deployment", latestDeploymentUID.String())
		}

		if err := s.metadataStore.DeleteDeployment(latestDeploymentUID.String()); err != nil {
			slog.Info("cannot delete current deployment of endpoint", "endpoint", endpointID, "deployment", latestDeploymentUID.String(), "msg", err.Error())
			return utils.WriteJSON(w, http.StatusBadRequest, utils.MakeErrorResponse(err))
		}
		if err := s.metadataStore.UpdateActiveDeploymentOfEndpoint(endpointID, currentDeploymentUID.String()); err != nil {
			slog.Info("cannot update active deployment of endpoint", "endpoint", endpointID, "deployment", currentDeploymentUID.String(), "msg", err.Error())
			return utils.WriteJSON(w, http.StatusBadRequest, utils.MakeErrorResponse(err))
		}
	default:
		_, err := s.metadataStore.GetDeploymentByID(deploymentID)
		if err != nil {
			slog.Info("cannot get given deploymentID", "endpoint", endpointID, "deployment", deploymentID, "msg", err.Error())
			return utils.WriteJSON(w, http.StatusBadRequest, utils.MakeErrorResponse(err))
		}
		// since deployments already in ascending order, we will loop from the beginning then check
		// if we meet the deployment then, we will delete following deployments

		sinceIdx := -1
		for idx, deployment := range deployments {
			if deployment.ID.String() == deploymentID {
				sinceIdx = idx + 1
				break
			}
		}

		if sinceIdx == -1 {
			return utils.WriteJSON(
				w,
				http.StatusBadRequest,
				map[string]any{
					"error": "unexpected error, could not found deployment from endpoint's deployments",
				})
		}

		for idx := sinceIdx; idx < len(deployments); idx++ {
			deploymentUID := deployments[idx].ID.String()
			blobMetadata, err := s.metadataStore.GetBlobMetadataByDeploymentID(deploymentUID)
			if err != nil {
				slog.Info("cannot get blob metadata", "endpoint", endpointID, "deployment", deploymentUID, "msg", err.Error())
				return utils.WriteJSON(w, http.StatusBadRequest, utils.MakeErrorResponse(err))
			}
			deleted, err := s.blobStore.DeleteDeploymentBlob(blobMetadata.Location)
			if err != nil {
				slog.Info("cannot remove blob from blob storage", "endpoint", endpointID, "deployment", deploymentUID, "msg", err.Error())
				return utils.WriteJSON(w, http.StatusBadRequest, utils.MakeErrorResponse(err))
			}

			if !deleted {
				slog.Info("delete blob failed", "endpoint", endpointID, "deployment", deploymentUID)
			}

			if err := s.metadataStore.DeleteDeployment(deploymentUID); err != nil {
				slog.Info("cannot delete given deployment of endpoint", "endpoint", endpointID, "deployment", deploymentUID, "msg", err.Error())
				return utils.WriteJSON(w, http.StatusBadRequest, utils.MakeErrorResponse(err))
			}
		}

		// rollback to specific deployment
		if err := s.metadataStore.UpdateActiveDeploymentOfEndpoint(endpointID, deploymentID); err != nil {
			slog.Info("cannot update active deployment of endpoint", "endpoint", endpointID, "deployment", deploymentID, "msg", err.Error())
			return utils.WriteJSON(w, http.StatusBadRequest, utils.MakeErrorResponse(err))
		}
	}
	return utils.WriteJSON(w, http.StatusOK, nil)
}

func handleStatus(w http.ResponseWriter, _ *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	status := map[string]string{"status": "ok"}
	return json.NewEncoder(w).Encode(status)
}

func FromInternalDeployment(d *types.Deployment) map[string]string {
	return map[string]string{
		"id":         d.ID.String(),
		"hash":       d.Hash,
		"endpointID": d.EndpointID.String(),
		"createdAt":  time.Unix(d.CreatedAt, 0).String(),
	}
}

type Endpoint struct {
	ID                 string              `json:"id,omitempty"`
	Name               string              `json:"name,omitempty"`
	Runtime            string              `json:"runtime,omitempty"`
	Environment        map[string]string   `json:"environment,omitempty"`
	ActiveDeploymentID string              `json:"activeDeploymentID,omitempty"`
	DeployHistory      []map[string]string `json:"deployHistory,omitempty"`
	CreatedAt          string              `json:"createdAt,omitempty"`
}

func FromInternalEndpoint(endpoint *types.Endpoint, deployments []*types.Deployment) Endpoint {
	var deployHistory []map[string]string
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
