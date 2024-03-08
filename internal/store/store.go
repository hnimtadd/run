package store

import (
	"github.com/hnimtadd/run/internal/types"

	"github.com/google/uuid"
)

type Store interface {
	CreateEndpoint(endpoint *types.Endpoint) error
	UpdateEndpoint(endpointID string, params UpdateEndpointParams) error
	GetEndpointByID(endpointID string) (*types.Endpoint, error)
	GetEndpoints() ([]*types.Endpoint, error)
	CreateDeployment(deploy *types.Deployment) error // TODO: should we mark this deployment as active deployment
	GetDeploymentByID(deploymentID string) (*types.Deployment, error)
	GetDeployments() ([]*types.Deployment, error)
}

type UpdateEndpointParams struct {
	Environment    map[string]string
	ActiveDeployID uuid.UUID
}
