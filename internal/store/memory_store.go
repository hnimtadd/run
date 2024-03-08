package store

import (
	"sync"

	"github.com/hnimtadd/run/internal/errors"
	"github.com/hnimtadd/run/internal/types"

	"github.com/google/uuid"
)

type MemoryStore struct {
	mu        sync.RWMutex
	deploys   map[uuid.UUID]*types.Deployment
	endpoints map[uuid.UUID]*types.Endpoint
}

func (m *MemoryStore) CreateEndpoint(endpoint *types.Endpoint) error {
	m.mu.Lock()
	_, existed := m.endpoints[endpoint.ID]
	m.mu.Unlock()
	if existed {
		return errors.ErrEndpointExisted
	}
	m.mu.RLock()
	m.endpoints[endpoint.ID] = endpoint
	m.mu.RUnlock()
	return nil
}

func (m *MemoryStore) UpdateEndpoint(endpointID string, params UpdateEndpointParams) error {
	endpointUUID, err := uuid.Parse(endpointID)
	if err != nil {
		return err
	}
	m.mu.Lock()
	curr, ok := m.endpoints[endpointUUID]
	m.mu.Unlock()
	if !ok {
		return errors.ErrEndpointNotExisted
	}
	curr.Environment = params.Environment
	curr.ActiveDeploymentID = params.ActiveDeployID
	m.mu.RLock()
	m.endpoints[endpointUUID] = curr
	m.mu.RUnlock()
	return nil
}

func (m *MemoryStore) GetEndpointByID(endpointID string) (*types.Endpoint, error) {
	endpointUUID, err := uuid.Parse(endpointID)
	if err != nil {
		return nil, err
	}

	endpoint, ok := m.endpoints[endpointUUID]
	if !ok {
		return nil, errors.ErrEndpointNotExisted
	}
	return endpoint, nil
}

func (m *MemoryStore) GetEndpoints() ([]*types.Endpoint, error) {
	var res []*types.Endpoint
	m.mu.Lock()
	for _, endpoint := range m.endpoints {
		res = append(res, endpoint)
	}
	m.mu.Unlock()
	return res, nil
}

func (m *MemoryStore) GetDeploymentByID(deploymentID string) (*types.Deployment, error) {
	uid, err := uuid.Parse(deploymentID)
	if err != nil {
		return nil, err
	}
	deployment, existed := m.deploys[uid]
	if !existed {
		return nil, errors.ErrDeploymentNotExisted
	}
	return deployment, nil
}

func (m *MemoryStore) GetDeployments() ([]*types.Deployment, error) {
	var res []*types.Deployment
	m.mu.Lock()
	for _, deployment := range m.deploys {
		res = append(res, deployment)
	}
	m.mu.Unlock()
	return res, nil
}

func (m *MemoryStore) CreateDeployment(deploy *types.Deployment) error {
	m.mu.Lock()
	_, existed := m.deploys[deploy.ID]
	m.mu.Unlock()
	if existed {
		return errors.ErrDeploymentExisted
	}
	m.mu.RLock()
	m.deploys[deploy.ID] = deploy
	m.endpoints[deploy.EndpointID].ActiveDeploymentID = deploy.ID
	m.mu.RUnlock()
	return nil
}

func (m *MemoryStore) GetDeployment(deploymentID string) (*types.Deployment, error) {
	deploymentUUID, err := uuid.Parse(deploymentID)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	deploy, ok := m.deploys[deploymentUUID]
	m.mu.Unlock()
	if !ok {
		return nil, errors.ErrDeploymentNotExisted
	}
	return deploy, nil
}

func NewMemoryStore() Store {
	return &MemoryStore{
		deploys:   make(map[uuid.UUID]*types.Deployment),
		endpoints: make(map[uuid.UUID]*types.Endpoint),
	}
}
