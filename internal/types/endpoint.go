package types

import (
	"time"

	"github.com/hnimtadd/run/internal/errors"

	"github.com/google/uuid"
)

var validRuntime = map[string]bool{
	"go": true,
}

func _isValidRuntime(runtime string) bool {
	return validRuntime[runtime]
}

type Endpoint struct {
	Environment        map[string]string `json:"environment"`
	Name               string            `json:"name"`
	Runtime            string            `json:"runtime"`
	CreatedAt          int64             `json:"createdAt"`
	ID                 uuid.UUID         `json:"id"`
	ActiveDeploymentID uuid.UUID         `json:"activeDeploymentId"`
}

func NewEndpoint(name string, runtime string, environment map[string]string) (*Endpoint, error) {
	if !_isValidRuntime(runtime) {
		return nil, errors.ErrInvalidRuntime
	}

	endpointID := uuid.New()
	endpoint := &Endpoint{
		ID:                 endpointID,
		Name:               name,
		CreatedAt:          time.Now().UnixMicro(),
		ActiveDeploymentID: uuid.Nil,
		Environment:        environment,
		Runtime:            runtime,
	}
	return endpoint, nil
}

func (e Endpoint) HasActiveDeploy() bool {
	return e.ID != uuid.Nil
}
