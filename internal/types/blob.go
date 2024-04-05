package types

import "github.com/google/uuid"

type Blob struct {
	ID           uuid.UUID `json:"id"`
	DeploymentID uuid.UUID `json:"deploymentID"`
	URL          string    `json:"url"`
	Data         []byte    `json:"data"`
	CreatedAt    int64     `json:"createdAt"`
}
