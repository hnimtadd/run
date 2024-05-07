package types

import (
	"time"

	"github.com/google/uuid"
)

type LogFormat byte

type RequestLog struct {
	RequestID    uuid.UUID `bson:"_id"`
	DeploymentID uuid.UUID `bson:"deployment_id"`
	Contents     []string  `bson:"contents"`
	CreatedAt    int64     `bson:"created_at"` // unix timestamp
}

func NewRequestLog(deploymentID uuid.UUID, requestID uuid.UUID, logs []string) *RequestLog {
	return &RequestLog{
		DeploymentID: deploymentID,
		RequestID:    requestID,
		Contents:     logs,
		CreatedAt:    time.Now().Unix(),
	}
}
