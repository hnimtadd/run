package store

import (
	"context"
	"sync"

	"github.com/hnimtadd/run/internal/errors"

	"github.com/google/uuid"
	"github.com/tetratelabs/wazero"
)

// ModCacher cache compiled module
type ModCacher interface {
	Put(deploymentID uuid.UUID, modCache wazero.CompilationCache) error
	Get(deploymentID uuid.UUID) (wazero.CompilationCache, error)
	Delete(deploymentID uuid.UUID) error
}

type MemoryModCacher struct {
	mu    sync.RWMutex
	cache map[uuid.UUID]wazero.CompilationCache
}

func (m *MemoryModCacher) Put(deploymentID uuid.UUID, modCache wazero.CompilationCache) error {
	m.mu.Lock()
	_, existed := m.cache[deploymentID]
	m.mu.Unlock()
	if existed {
		return errors.Newf("mod cacher: cannot put cache into store, %v", errors.ErrDeploymentExisted)
	}
	m.mu.RLock()
	m.cache[deploymentID] = modCache
	m.mu.RUnlock()
	return nil
}

func (m *MemoryModCacher) Get(deploymentID uuid.UUID) (wazero.CompilationCache, error) {
	m.mu.Lock()
	cache, existed := m.cache[deploymentID]
	m.mu.Unlock()
	if !existed {
		return nil, errors.Newf("mod cacher: cannot get cache, %v", errors.ErrDeploymentNotExisted)
	}
	return cache, nil
}

func (m *MemoryModCacher) Delete(deploymentID uuid.UUID) error {
	// if deploymentID not existed, still return
	m.mu.Lock()
	curr, existed := m.cache[deploymentID]
	m.mu.Unlock()
	if !existed {
		return errors.Newf("mod cacher: cannot delete cache, %v", errors.ErrDeploymentNotExisted)
	}
	delete(m.cache, deploymentID)
	if err := curr.Close(context.Background()); err != nil {
		return errors.Newf("mod cachder: cannot close module, %v", err)
	}
	return nil
}

func NewMemoryModCacher() ModCacher {
	return &MemoryModCacher{
		cache: make(map[uuid.UUID]wazero.CompilationCache),
	}
}
