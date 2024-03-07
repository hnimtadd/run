package store

import (
	"github.com/google/uuid"
	"github.com/hnimtadd/run/internal/types"
)

type Store interface {
	CreateEndpoint(endpoint *types.Endpoint) error
	UpdateEndpoint(endpointID string, params UpdateEndpointParams) error
	GetEndpointByID(endpointID string) (*types.Endpoint, error)
	GetEndpoints() ([]*types.Endpoint, error)
	CreateDeployment(deploy *types.Deployment) error
	GetDeploymentByID(deploymentID string) (*types.Deployment, error)
	GetDeployments() ([]*types.Deployment, error)
}

type UpdateEndpointParams struct {
	Environment    map[string]string
	ActiveDeployID uuid.UUID
}
