package store

import (
	"github.com/google/uuid"
	"github.com/hnimtadd/run/internal/types"
)

type Store interface {
	CreateEndpoint(endpoint *types.Endpoint) error
	UpdateEndpoint(endpointID string, params UpdateEndpointParams) error
	GetEndpoint(endpointID string) (*types.Endpoint, error)
	CreateDeployment(deploy *types.Deployment) error
	GetDeployment(deploymentID string) (*types.Deployment, error)
}

type UpdateEndpointParams struct {
	Environment    map[string]string
	ActiveDeployID uuid.UUID
}
