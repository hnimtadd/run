package store

import (
	"github.com/hnimtadd/run/internal/types"

	"github.com/google/uuid"
)

type (
	Store interface {
		CreateEndpoint(endpoint *types.Endpoint) error
		UpdateEndpoint(endpointID string, params UpdateEndpointParams) error
		GetEndpointByID(endpointID string) (*types.Endpoint, error)
		GetEndpoints() ([]*types.Endpoint, error)
		UpdateActiveDeploymentOfEndpoint(endpointID string, deploymentID string) error

		CreateDeployment(deploy *types.Deployment) error
		GetDeploymentByID(deploymentID string) (*types.Deployment, error)
		GetDeployments() ([]*types.Deployment, error)

		GetDeploymentByEndpointID(endpointID string) ([]*types.Deployment, error)
	}
	UpdateEndpointParams struct {
		Environment    map[string]string
		ActiveDeployID uuid.UUID
	}
	LogStore interface {
		AppendLog(log *types.RequestLog) error
		GetLogByRequestID(requestID string) (*types.RequestLog, error)
		GetLogsOfRequest(deploymentID string, requestID string) (*types.RequestLog, error)
		GetLogOfDeployment(deploymentID string) ([]*types.RequestLog, error)
	}
)
