package store

import (
	"sync"
	"time"

	"github.com/hnimtadd/run/internal/errors"
	"github.com/hnimtadd/run/internal/types"

	"github.com/google/uuid"
)

var (
	_ LogStore = &MemoryStore{}
	_ Store    = &MemoryStore{}
)

type MemoryStore struct {
	mu        sync.RWMutex
	deploys   map[uuid.UUID]*types.Deployment
	endpoints map[uuid.UUID]*types.Endpoint
	logs      map[uuid.UUID]map[uuid.UUID]*types.RequestLog // map deploymentID with request_id and request_log.go
}

func (m *MemoryStore) UpdateActiveDeploymentOfEndpoint(endpointID string, deploymentID string) error {
	endpointUID, err := uuid.Parse(endpointID)
	if err != nil {
		return err
	}
	deploymentUID, err := uuid.Parse(deploymentID)
	if err != nil {
		return err
	}

	m.mu.Lock()
	_, ok := m.endpoints[endpointUID]
	m.mu.Unlock()
	if !ok {
		return errors.ErrDocumentNotFound
	}
	m.endpoints[endpointUID].ActiveDeploymentID = deploymentUID
	return nil
}

func (m *MemoryStore) AppendLog(log *types.RequestLog) error {
	m.mu.Lock()
	_, ok := m.logs[log.DeploymentID]
	m.mu.Unlock()

	if !ok {
		m.mu.RLock()
		m.logs[log.DeploymentID] = make(map[uuid.UUID]*types.RequestLog)
		m.mu.RUnlock()
	}

	m.mu.Lock()
	_, ok = m.logs[log.DeploymentID][log.RequestID]
	m.mu.Unlock()

	if ok {
		return errors.ErrDocumentDuplicated
	}

	m.mu.RLock()
	m.logs[log.DeploymentID][log.RequestID] = log
	m.mu.RUnlock()
	return nil
}

func (m *MemoryStore) GetLogOfDeployment(deploymentID string) ([]*types.RequestLog, error) {
	deploymentUUID, err := uuid.Parse(deploymentID)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	entry, ok := m.logs[deploymentUUID]
	m.mu.Unlock()
	if !ok {
		return nil, errors.ErrDeploymentNotExisted
	}

	var eventLogs []*types.RequestLog
	for _, eventLog := range entry {
		eventLogs = append(eventLogs, eventLog)
	}
	return eventLogs, nil
}

func (m *MemoryStore) GetLogsOfRequest(deploymentID string, requestID string) (*types.RequestLog, error) {
	deploymentUUID, err := uuid.Parse(deploymentID)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	entry, ok := m.logs[deploymentUUID]
	m.mu.Unlock()
	if !ok {
		return nil, errors.ErrDeploymentNotExisted
	}

	requestUUID, err := uuid.Parse(requestID)
	if err != nil {
		return nil, err
	}

	log, ok := entry[requestUUID]
	if !ok {
		return nil, errors.ErrDocumentNotFound
	}
	return log, nil
}

func (m *MemoryStore) GetLogByRequestID(requestID string) (*types.RequestLog, error) {
	requestUUID, err := uuid.Parse(requestID)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	for _, entry := range m.logs {
		log, ok := entry[requestUUID]
		if ok {
			return log, nil
		}
	}
	return nil, errors.ErrDocumentNotFound
}

func (m *MemoryStore) CreateEndpoint(endpoint *types.Endpoint) error {
	m.mu.Lock()
	_, existed := m.endpoints[endpoint.ID]
	m.mu.Unlock()
	if existed {
		return errors.ErrEndpointExisted
	}
	now := time.Now()
	endpoint.CreatedAt = now.Unix()
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
	m.mu.Lock()
	deployment, existed := m.deploys[uid]
	m.mu.Unlock()
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

func (m *MemoryStore) GetDeploymentByEndpointID(endpointID string) ([]*types.Deployment, error) {
	endpointUID, err := uuid.Parse(endpointID)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	var deployments []*types.Deployment
	for _, deployment := range m.deploys {
		if deployment.EndpointID == endpointUID {
			deployments = append(deployments, deployment)
		}
	}
	m.mu.Unlock()

	return deployments, nil
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		deploys:   make(map[uuid.UUID]*types.Deployment),
		endpoints: make(map[uuid.UUID]*types.Endpoint),
		logs:      make(map[uuid.UUID]map[uuid.UUID]*types.RequestLog),
	}
}
