package types

import (
	"time"

	"github.com/google/uuid"
)

type LogFormat byte

type RequestLog struct {
	RequestID    uuid.UUID
	DeploymentID uuid.UUID
	Contents     []string
	CreatedAt    int64 // unix timestamp
}

func NewRequestLog(deploymentID uuid.UUID, requestID uuid.UUID, logs []string) *RequestLog {
	return &RequestLog{
		DeploymentID: deploymentID,
		RequestID:    requestID,
		Contents:     logs,
		CreatedAt:    time.Now().Unix(),
	}
}
