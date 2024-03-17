package types

import (
	"crypto/md5"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

type Deployment struct {
	Hash       string    `json:"hash"`
	Blob       []byte    `json:"blob"`
	CreatedAt  int64     `json:"createdAt"`
	ID         uuid.UUID `json:"id"`
	EndpointID uuid.UUID `json:"endpointID"`
	Format     LogFormat `json:"logFormat"`
}

func NewDeployment(endpoint *Endpoint, blob []byte) (*Deployment, error) {
	md5hash := md5.Sum(blob)
	deploymentHash := hex.EncodeToString(md5hash[:])
	deploymentID := uuid.New()
	deployment := &Deployment{
		ID:         deploymentID,
		Blob:       blob,
		Hash:       deploymentHash,
		EndpointID: endpoint.ID,
		CreatedAt:  time.Now().UnixMicro(),
	}
	return deployment, nil
}
